/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/string"
	dockerclient "github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"
)

type Bootstrapper interface {
	Bootstrap(ctx context.Context, sb sandboxer.Sandbox, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error)
}

type Config struct {
	Implementation string `toml:"implementation" json:"implementation"`
	Endpoint       string `toml:"endpoint" json:"endpoint,omitempty"`
	Image          string `toml:"image" json:"image,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Implementation: "docker",
		Image:          "ghcr.io/drassi-run/ubuntu:26.04",
	}
}

type engine struct {
	client       container.Engine
	defaultImage string
}

func New(config *Config) (sandboxer.Engine, error) {
	if config.Implementation != "docker" {
		return nil, fmt.Errorf("unsupported container implementation: %s", config.Implementation)
	}

	opts := make([]dockerclient.Opt, 0)
	if ep := config.Endpoint; ep != "" {
		opt := dockerclient.WithHost(ep)
		opts = append(opts, opt)
	}

	client, err := docker.New(opts...)
	if err != nil {
		return nil, err
	}
	client = container.WithTelemetry(client)

	e := &engine{
		client:       client,
		defaultImage: config.Image,
	}
	return e, nil
}

func NewBootstrapper(client container.Engine) Bootstrapper {
	return &engine{client: client}
}

func (e *engine) Close() error {
	return e.client.Close()
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	var (
		sb          sandboxer.Sandbox
		containerId string
	)

	if req.JobContainer == nil {
		spec := &types.ContainerSpec{
			Image:      e.defaultImage,
			Entrypoint: []string{"sleep"},
			Command:    []string{"infinity"},
		}
		spec.NetworkMode = "host"
		runOpts := &container.RunOptions{
			Stdio:   new(types.Stdio),
			Streams: new(stream.Streams),
		}
		if cid, err := e.client.ContainerRun(ctx, spec, runOpts); err != nil {
			return nil, err
		} else if sb, err = newSandbox(ctx, e.client, cid); err != nil {
			return nil, err
		} else {
			containerId = cid
		}
	}

	resp, err := e.Bootstrap(ctx, sb, req)
	if err != nil {
		return nil, err
	}
	if containerId != "" {
		resp.JobContainer = &records.ContainerInfo{
			Id: containerId,
		}
	}
	return resp, nil
}

func (e *engine) Bootstrap(ctx context.Context, sb sandboxer.Sandbox, req *sandboxer.LaunchRequest) (resp *sandboxer.LaunchResponse, err error) {
	resp = &sandboxer.LaunchResponse{
		Sandbox:         sb,
		ContainerEngine: e.client,
	}

	if req.JobContainer == nil && len(req.ServiceContainers) == 0 {
		if sb == nil {
			return nil, fmt.Errorf("NIL sandbox")
		}
		return resp, nil
	}

	labels := types.LabelsFor(req.Github)
	// cleanup order is matter
	cleanups := []sandboxer.Cleanup{
		cleanup(labels, e.client.ContainerRemove),
		cleanup(labels, e.client.VolumeRemove),
		cleanup(labels, e.client.NetworkRemove),
	}

	// Create network for job container, services containers and all container actions
	networkId, err := e.client.NetworkCreate(ctx, &types.NetworkSpec{
		Name:   e.nameFor(req.Github),
		Driver: "bridge",
		Labels: labels,
	})
	if err != nil {
		return nil, err
	}

	// Run job container
	if def := req.JobContainer; def != nil {
		refiners := []refiner{
			setCmd([]string{"sleep"}, []string{"infinity"}),
			setWorkdir(defaultLayout.Workspace),
			setNetwork(networkId),
			addSandboxMounts(sb),
			addContainerSocketMounts(e.client),
			setLabels(labels),
			setCIEnv(),
		}
		containerId, err := e.runContainer(ctx, def, refiners)
		if err != nil {
			return nil, err
		}
		resp.JobContainer = &records.ContainerInfo{
			Id:      containerId,
			Network: networkId,
		}

		if sb, err = newSandbox(ctx, e.client, containerId); err != nil {
			return nil, err
		}
		sb = sandboxer.AddAfterCleanup(sb, cleanups...)
		if resp.Sandbox != nil {
			resp.Sandbox = sandboxer.NewLayeredSandbox(sb, resp.Sandbox)
		}
	} else {
		resp.Sandbox = sandboxer.AddBeforeCleanup(resp.Sandbox, cleanups...)
	}

	// Run services container in parallel
	if len(req.ServiceContainers) > 0 {
		refiners := []refiner{
			setNetwork(networkId),
			setLabels(labels),
		}
		g, ctx := errgroup.WithContext(ctx)
		g.SetLimit(8)
		for name, def := range req.ServiceContainers {
			name, def := name, def
			g.Go(func() error {
				if containerId, err := e.runContainer(ctx, def, refiners); err != nil {
					return err
				} else if portMap, err := e.getPortsMap(ctx, containerId); err != nil {
					return err
				} else {
					resp.ServiceContainers[name] = &records.ContainerInfo{
						Id:      containerId,
						Network: networkId,
						Ports:   portMap,
					}
					return nil
				}
			})
		}
		if err = g.Wait(); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func (e *engine) parseContainer(def *workflows.Container, refiners []refiner) (spec *types.ContainerSpec, err error) {
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
		if vol, err := cli.ParseVolume(v); err != nil {
			return nil, err
		} else {
			spec.Mounts = append(spec.Mounts, vol)
		}
	}
	for _, p := range def.Ports {
		if pb, length, err := cli.ParsePublish(p); err != nil {
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

	for _, fn := range refiners {
		if err = fn(spec); err != nil {
			return nil, err
		}
	}

	return
}

func (e *engine) runContainer(ctx context.Context, def *workflows.Container, refiners []refiner) (string, error) {
	spec, err := e.parseContainer(def, refiners)
	if err != nil {
		return "", err
	}

	pullOpts := &container.PullOptions{
		PullPolicy: spec.PullPolicy,
	}
	if cred := def.Credentials; cred != nil {
		pullOpts.RegistryAuth = container.NewBasicAuth(cred.Username, cred.Password)
	}
	if err = e.client.ImagePull(ctx, def.Image, pullOpts); err != nil {
		return "", err
	}

	runOpts := &container.RunOptions{
		Stdio:   new(types.Stdio),
		Streams: new(stream.Streams),
	}
	return e.client.ContainerRun(ctx, spec, runOpts)
}

func (e *engine) nameFor(gh *records.Github) string {
	repo := xstring.Normalize(gh.Repository)
	repo = strings.ToLower(repo)

	workflow := strings.TrimSuffix(gh.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(gh.Job)
	run := xstring.Normalize(gh.RunId)
	attempt := xstring.Normalize(gh.RunAttempt)

	name := strings.Join([]string{repo, workflow, job, run, attempt}, "-")
	return name
}

func (e *engine) getPortsMap(ctx context.Context, id string) (map[string]string, error) {
	spec, err := e.client.ContainerInspect(ctx, id)
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
