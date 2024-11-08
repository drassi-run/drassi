package runtime

import (
	"context"
	"maps"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	. "drassi.run/core/util/types"
)

// Container runtime is used to run docker action
type Container interface {
	// TranslatePath map from containerPath to sandboxPath,
	TranslatePath(containerPath string) (sandboxPath string, ok bool)

	Pull(ctx context.Context, image string, auth container.RegistryAuth) error
	Build(ctx context.Context) error
	Run(ctx context.Context, image string, entrypoint, cmd []string, env map[string]string) error
}

type containerRuntime struct {
	engine  container.Engine
	streams container.Streams
	network string
	mounts  []Pair[string, *types.Mount] // list of (sandboxPath, *Mount) pair, sorted by sandboxPath
	pathMap [][2]string                  // list of (containerPath, sandboxPath) pair, sorted by containerPath
}

func (rt *containerRuntime) TranslatePath(containerPath string) (sandboxPath string, ok bool) {
	sandboxPath = mapPath(containerPath, rt.pathMapSeq)
	ok = sandboxPath != ""
	return
}

func (rt *containerRuntime) pathMapSeq(yield func(string, string) bool) {
	for _, pair := range rt.pathMap {
		if !yield(pair[0], pair[1]) {
			return
		}
	}
}

func (rt *containerRuntime) Pull(ctx context.Context, image string, auth container.RegistryAuth) error {
	return rt.engine.ImagePull(ctx, image, &container.PullOptions{
		RegistryAuth: auth,
		Streams:      rt.streams,
	})
}

func (rt *containerRuntime) Build(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (rt *containerRuntime) Run(ctx context.Context, image string, entrypoint, cmd []string, env map[string]string) error {
	// clone env to avoid modify the original
	runEnv := maps.Clone(env)
	for k, v := range env {
		if path := mapPath(v, rt.pathMapSeq); path != "" {
			runEnv[k] = path
		}
	}

	spec := &types.ContainerSpec{
		Image:       image,
		Entrypoint:  entrypoint,
		Command:     cmd,
		Environment: runEnv,
	}
	if rt.network != "" {
		ep := &types.Endpoint{
			Target: rt.network,
		}
		spec.Endpoints = append(spec.Endpoints, ep)
	}
	for _, mount := range rt.mounts {
		spec.Mounts = append(spec.Mounts, mount.Value)
	}

	stdio := new(types.Stdio)
	stdio.Attach = types.Stdout | types.Stderr
	_, err := rt.engine.ContainerRun(ctx, spec, &container.RunOptions{
		Stdio:   stdio,
		Streams: rt.streams,
	})
	return err
}

func (rt *containerRuntime) mountMapSeq(yield func(string, string) bool) {
	for _, mount := range rt.mounts {
		if !yield(mount.Key, mount.Value.Target) {
			return
		}
	}
}
