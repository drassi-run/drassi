/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"errors"
	"maps"
	"strconv"
	"sync"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/parser"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
	xsync "drassi.run/core/util/sync"
	"golang.org/x/sync/errgroup"
)

func NewBootstrapper(client container.Engine, forge *records.Forge) *Bootstrapper {
	labels := forge.WellKnownLabels()
	return &Bootstrapper{
		client: client,
		forge:  forge,
		labels: labels,
		cleanups: []sandboxer.Cleanup{
			removeByLabels(labels, client.ContainerRemove),
			removeByLabels(labels, client.VolumeRemove),
		},
	}
}

// Bootstrapper coordinates the lifecycle of job and service containers for a sandbox,
// including network provisioning, container execution, and cleanup.
//
// Bootstrapper is NOT thread-safe.
type Bootstrapper struct {
	client   container.Engine
	forge    *records.Forge
	labels   map[string]string
	cleanups []sandboxer.Cleanup

	createNetwork func(context.Context) (string, error)
}

func (b *Bootstrapper) Network(id string, ephemeral bool) {
	b.createNetwork = func(context.Context) (string, error) {
		return id, nil
	}
	if ephemeral {
		b.cleanups = append(b.cleanups, func(ctx context.Context) error {
			return b.client.NetworkRemove(ctx, &container.RemoveOptions{Id: id})
		})
	}
}

func (b *Bootstrapper) ensureNetwork(ctx context.Context) (string, error) {
	if b.createNetwork == nil {
		b.createNetwork = xsync.Singleton(b.doCreateNetwork)
	}
	return b.createNetwork(ctx)
}

func (b *Bootstrapper) doCreateNetwork(ctx context.Context) (string, error) {
	netId, err := b.client.NetworkCreate(ctx, &types.NetworkSpec{
		Name:   b.forge.CanonicalName(),
		Driver: "bridge",
		Labels: b.labels,
	})
	if err != nil {
		return "", err
	}
	b.cleanups = append(b.cleanups, func(ctx context.Context) error {
		return b.client.NetworkRemove(ctx, &container.RemoveOptions{Id: netId})
	})
	return netId, nil
}

func (b *Bootstrapper) RunJobContainer(ctx context.Context, def *workflows.Container, opts ...Option) (*records.ContainerInfo, error) {
	if def == nil {
		return nil, nil
	}

	netId, err := b.ensureNetwork(ctx)
	if err != nil {
		return nil, err
	}

	layout := DefaultLayout
	options := []Option{
		SetNetwork(netId),
		SetLabels(b.labels),
		SetWorkdir(layout.Workspace()),
		SetCIEnv(),
		SetCmd([]string{"sleep"}, []string{"infinity"}),
		MountApiSocket(b.client),
	}
	options = append(options, opts...)

	conId, err := b.runContainer(ctx, def, options...)
	if err != nil {
		return nil, err
	}

	return &records.ContainerInfo{
		Id:      conId,
		Network: netId,
	}, nil
}

func (b *Bootstrapper) RunServiceContainers(ctx context.Context, defs map[string]*workflows.Container, opts ...Option) (map[string]*records.ContainerInfo, error) {
	if len(defs) == 0 {
		return make(map[string]*records.ContainerInfo), nil
	}

	netId, err := b.ensureNetwork(ctx)
	if err != nil {
		return nil, err
	}

	options := []Option{
		SetNetwork(netId),
		SetLabels(b.labels),
	}
	options = append(options, opts...)

	var mu sync.Mutex
	res := make(map[string]*records.ContainerInfo, len(defs))

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(8)

	for name, def := range defs {
		g.Go(func() error {
			conId, err := b.runContainer(gCtx, def, options...)
			if err != nil {
				return err
			}

			portMap, err := b.getPortsMap(gCtx, conId)
			if err != nil {
				return err
			}

			mu.Lock()
			res[name] = &records.ContainerInfo{
				Id:      conId,
				Network: netId,
				Ports:   portMap,
			}
			mu.Unlock()
			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (b *Bootstrapper) WrapSandbox(underlay, overlay sandboxer.Sandbox) sandboxer.Sandbox {
	if overlay == nil && underlay == nil {
		panic("both sandbox layers are nil")
	}

	if overlay == nil {
		return sandboxer.AddBeforeCleanup(underlay, b.cleanups...)
	}

	overlay = sandboxer.AddAfterCleanup(overlay, b.cleanups...)
	if underlay == nil { // in container-native sandboxer, e.g. docker
		return overlay
	}

	return sandboxer.NewLayeredSandbox(overlay, underlay)
}

func (b *Bootstrapper) Rollback(ctx context.Context) error {
	var errs []error
	for _, fn := range b.cleanups {
		if err := fn(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b *Bootstrapper) runContainer(ctx context.Context, def *workflows.Container, opts ...Option) (string, error) {
	spec, err := b.parseContainer(def, opts...)
	if err != nil {
		return "", err
	}

	pullOpts := &container.PullOptions{
		PullPolicy: spec.PullPolicy,
	}
	if cred := def.Credentials; cred != nil {
		pullOpts.RegistryAuth = container.NewBasicAuth(cred.Username, cred.Password)
	}
	if err = b.client.ImagePull(ctx, def.Image, pullOpts); err != nil {
		return "", err
	}

	runOpts := &container.RunOptions{
		Stdio:   new(types.Stdio),
		Streams: new(stream.Streams),
	}
	return b.client.ContainerRun(ctx, spec, runOpts)
}

func (b *Bootstrapper) parseContainer(def *workflows.Container, opts ...Option) (spec *types.ContainerSpec, err error) {
	if spec, _, err = cli.Parse(def.Options); err != nil {
		return
	}

	spec.Image = def.Image
	if env := def.Env; len(env) > 0 {
		if spec.Environment == nil {
			spec.Environment = make(map[string]string)
		}
		maps.Copy(spec.Environment, env)
	}
	for _, v := range def.Volumes {
		if vol, err := parser.ParseVolume(v); err != nil {
			return nil, err
		} else {
			spec.Mounts = append(spec.Mounts, vol)
		}
	}
	for _, p := range def.Ports {
		if pb, length, err := parser.ParsePublish(p); err != nil {
			return nil, err
		} else {
			for i := range length {
				binding := &types.PortBinding{HostIP: pb.HostIP, ContainerPort: pb.ContainerPort + i, Protocol: pb.Protocol}
				if binding.HostPort != 0 {
					binding.HostPort = binding.HostPort + i
				}
				spec.Publish = append(spec.Publish, binding)
			}
		}
	}

	for _, fn := range opts {
		if err = fn(spec); err != nil {
			return nil, err
		}
	}

	return
}

func (b *Bootstrapper) getPortsMap(ctx context.Context, id string) (map[string]string, error) {
	spec, err := b.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}
	portMap := make(map[string]string)
	for _, pb := range spec.Publish {
		if pb.Protocol != "tcp" {
			continue
		}
		containerPort := strconv.FormatUint(uint64(pb.ContainerPort), 10)
		hostPort := strconv.FormatUint(uint64(pb.HostPort), 10)
		portMap[containerPort] = hostPort
	}
	return portMap, nil
}
