package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	. "drassi.run/core/util/types"
	"k8s.io/utils/set"
)

// Container runtime is used to run docker action
type Container interface {
	PathTranslator

	Pull(ctx context.Context, image string, auth container.RegistryAuth) error
	Build(ctx context.Context) error
	Run(ctx context.Context, image string, entrypoint, cmd []string, env map[string]string) error
}

type containerRuntime struct {
	engine  container.Engine
	streams container.Streams

	workdir string
	labels  map[string]string
	network string
	mounts  []Pair[string, *types.Mount] // list of (sandboxPath, *Mount) pair, sorted DESC by sandboxPath
	pathMap [][2]string                  // list of (containerPath, sandboxPath) pair, sorted DESC by containerPath
}

type ContainerRuntimeOption func(*containerRuntime)

func WithWorkDir(workdir string) ContainerRuntimeOption {
	return func(rt *containerRuntime) {
		rt.workdir = workdir
	}
}

func WithLabels(labels map[string]string) ContainerRuntimeOption {
	return func(rt *containerRuntime) {
		rt.labels = labels
	}
}

func WithNetwork(network string) ContainerRuntimeOption {
	return func(rt *containerRuntime) {
		rt.network = network
	}
}

func WithMounts(mounts []Pair[string, *types.Mount]) ContainerRuntimeOption {
	return func(rt *containerRuntime) {
		rt.mounts = append(rt.mounts, mounts...)
	}
}

func NewContainerRuntime(engine container.Engine, streams container.Streams, opts ...ContainerRuntimeOption) (Container, error) {
	rt := &containerRuntime{
		engine:  engine,
		streams: streams,
	}
	for _, o := range opts {
		o(rt)
	}

	containerPaths := set.New[string]()
	sandboxerPaths := set.New[string]()
	mounts := rt.mounts
	pathMap := make([][2]string, 0, len(mounts))
	for _, m := range mounts {
		sPath, cPath := m.Key, m.Value.Target
		if s := strings.TrimRight(sPath, "/"); sandboxerPaths.Has(s) {
			return nil, fmt.Errorf("found duplicate sandbox mount at %q", s)
		} else {
			sandboxerPaths.Insert(s)
		}
		if s := strings.TrimRight(cPath, "/"); containerPaths.Has(s) {
			return nil, fmt.Errorf("found duplicate container mount at %q", s)
		} else {
			containerPaths.Insert(s)
		}
		pathMap = append(pathMap, [...]string{cPath, sPath})
	}

	slices.SortFunc(mounts, func(a, b Pair[string, *types.Mount]) int {
		return strings.Compare(b.Key, a.Key) // DESC order
	})
	slices.SortFunc(pathMap, func(a, b [2]string) int {
		return strings.Compare(b[0], a[0]) // DESC order
	})
	rt.mounts = mounts
	rt.pathMap = pathMap

	return rt, nil
}

// TranslatePath map from containerPath to sandboxPath,
func (rt *containerRuntime) TranslatePath(containerPath string) (sandboxPath string, ok bool) {
	sandboxPath = MapPath(containerPath, rt.pathMapSeq)
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
		if path := MapPath(v, rt.mountMapSeq); path != "" {
			runEnv[k] = path
		}
	}

	spec := &types.ContainerSpec{
		Image:       image,
		Entrypoint:  entrypoint,
		Command:     cmd,
		Environment: runEnv,
		WorkingDir:  rt.workdir,
		Labels:      rt.labels,
	}
	spec.AutoRemove = true
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
	for _, m := range rt.mounts {
		if !yield(m.Key, m.Value.Target) {
			return
		}
	}
}
