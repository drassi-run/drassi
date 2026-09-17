/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"context"
	"fmt"
	"maps"
	"sync/atomic"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/container/parser"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/stream"
	dockerclient "github.com/moby/moby/client"
)

func New(cfg *Config, prov *provision.Provisioner[*types.ContainerSpec]) (sandboxer.Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	opts := make([]dockerclient.Opt, 0)
	if ep := cfg.Endpoint; ep != "" {
		opts = append(opts, dockerclient.WithHost(ep))
	}

	client, err := docker.New(opts...)
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	e := &engine{
		client:      client,
		template:    cfg.Template,
		provisioner: prov,
	}
	return e, nil
}

type engine struct {
	client      c.Engine
	template    *container.Template
	provisioner *provision.Provisioner[*types.ContainerSpec]
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (resp *sandboxer.LaunchResponse, err error) {
	labels := req.Forge.WellKnownLabels()
	netId, err := e.client.NetworkCreate(ctx, &types.NetworkSpec{
		Name:   req.Forge.CanonicalName(),
		Driver: "bridge",
		Labels: labels,
	})
	if err != nil {
		return
	}

	b := container.NewBootstrapper(e.client, req.Forge)
	b.Network(netId, true)

	var sb sandboxer.Sandbox
	defer func(ctx context.Context) {
		if err == nil {
			return
		}
		ctx = context.WithoutCancel(ctx)
		if sb != nil {
			_ = sb.Terminate(ctx)
		}
		_ = b.Rollback(ctx)
		resp = nil
	}(ctx)

	volId, err := e.client.VolumeCreate(ctx, &types.VolumeSpec{
		Name:   req.Forge.CanonicalName(),
		Labels: labels,
		Driver: "local",
	})
	if err != nil {
		return
	}

	layout := container.DefaultLayout
	spec, err := e.resolveTemplate(req,
		container.MountVolume(volId, container.DefaultJobDir),
		container.MountApiSocket(e.client),
		container.SetNetwork(netId),
		container.SetCIEnv(),
		container.SetWorkdir(layout.Workspace()),
		container.SetCmd([]string{"sleep"}, []string{"infinity"}),
		container.SetLabels(labels),
	)
	if err != nil {
		return
	}

	id := new(atomic.Value)
	launcher := e.launch(layout, id)
	if prov := e.provisioner; prov != nil {
		launcher = prov.Launch(layout.Runtimes(), launcher)
	}

	sb, err = launcher(ctx, spec)
	if err != nil {
		return
	}

	conId, ok := id.Load().(string)
	if !ok || conId == "" {
		return nil, fmt.Errorf("sandbox NOT running")
	}

	resp = &sandboxer.LaunchResponse{
		Mounter: newMounter(volId),
		ContainerEngine: func(context.Context) (c.Engine, error) {
			return e.client, nil
		},
		JobContainer: &records.ContainerInfo{
			Id:      conId,
			Network: netId,
		},
	}

	if len(req.ServiceContainers) > 0 {
		if resp.ServiceContainers, err = b.RunServiceContainers(ctx, req.ServiceContainers); err != nil {
			return
		}
	}

	resp.Sandbox = b.WrapSandbox(nil, sb)
	return
}

func (e *engine) launch(layout sandboxer.Layout, id *atomic.Value) provision.Launcher[*types.ContainerSpec] {
	return func(ctx context.Context, spec *types.ContainerSpec) (sandboxer.Sandbox, error) {
		runOpts := &c.RunOptions{
			Stdio:   new(types.Stdio),
			Streams: new(stream.Streams),
		}
		if cid, err := e.client.ContainerRun(ctx, spec, runOpts); err != nil {
			return nil, err
		} else if old := id.Swap(cid); old != nil && old != "" {
			return nil, fmt.Errorf("sandbox already running: %s", old)
		} else {
			return container.NewSandbox(ctx, e.client, cid, layout)
		}
	}
}

func (e *engine) resolveTemplate(req *sandboxer.LaunchRequest, opts ...container.Option) (*types.ContainerSpec, error) {
	tmpl := e.template
	if tmpl == nil {
		tmpl = &container.Template{Image: container.DefaultImage}
	}

	spec := tmpl.ContainerSpec()
	if job := req.JobContainer; job != nil {
		spec.Image = job.Image

		if len(job.Env) > 0 {
			if spec.Environment == nil {
				spec.Environment = make(map[string]string)
			}
			maps.Copy(spec.Environment, job.Env)
		}

		for _, v := range job.Volumes {
			if vol, err := parser.ParseVolume(v); err == nil {
				spec.Mounts = append(spec.Mounts, vol)
			}
		}

		for _, p := range job.Ports {
			if pb, length, err := parser.ParsePublish(p); err == nil {
				for i := range length {
					binding := &types.PortBinding{
						HostIP:        pb.HostIP,
						ContainerPort: pb.ContainerPort + i,
						Protocol:      pb.Protocol,
					}
					if pb.HostPort != 0 {
						binding.HostPort = pb.HostPort + i
					}
					spec.Publish = append(spec.Publish, binding)
				}
			}
		}

		if s, _, err := cli.Parse(job.Options); err == nil && s != nil {
			if len(s.Mounts) > 0 {
				spec.Mounts = append(spec.Mounts, s.Mounts...)
			}
			if len(s.Publish) > 0 {
				spec.Publish = append(spec.Publish, s.Publish...)
			}
			if s.User != "" {
				spec.User = s.User
			}
			if s.Privileged {
				spec.Privileged = true
			}
		}
	}

	for _, o := range opts {
		if err := o(spec); err != nil {
			return nil, err
		}
	}
	return spec, nil
}

func (e *engine) Close() error {
	return e.client.Close()
}
