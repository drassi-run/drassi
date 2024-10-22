package container

import (
	"context"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
)

type engine struct {
	client       container.Engine
	defaultImage string
}

func New(spec *ContainerSpec) (sandboxer.Engine, error) {
	client, err := docker.New()
	if err != nil {
		return nil, err
	}

	e := &engine{
		client:       client,
		defaultImage: spec.Image,
	}
	return e, nil
}

func (e *engine) Close() error {
	return e.client.Close()
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	spec := &types.ContainerSpec{
		Image:       e.defaultImage,
		Entrypoint:  []string{"sleep"},
		Command:     []string{"infinity"},
		NetworkMode: "host",
	}
	runOpts := &container.RunOptions{
		Stdio: &types.Stdio{
			Tty:         false,
			Interactive: false,
			Attach:      types.None,
		},
		Streams: nil,
	}
	containerId, err := e.client.ContainerRun(ctx, spec, runOpts)
	if err != nil {
		return nil, err
	}
	sb, err := newSandbox(ctx, e.client, containerId)
	if err != nil {
		return nil, err
	}

	res := &sandboxer.LaunchResponse{
		Sandbox:         sb,
		ContainerEngine: e.client,
	}
	return res, nil
}
