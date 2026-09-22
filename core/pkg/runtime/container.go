/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/specdef"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/stream"
	"github.com/emirpasic/gods/v2/sets/hashset"
)

// Container runtime is used to run docker action
type Container interface {
	PathTranslator

	Pull(ctx context.Context, image string, auth container.RegistryAuth) error
	Build(ctx context.Context) error
	Run(ctx context.Context, image string, entrypoint, cmd []string, env map[string]string, streams *stream.Streams) error
}

type containerRuntime struct {
	engine container.Engine
	spec   *types.ContainerSpec

	mountMap [][2]string // list of (sandboxPath, containerPath) pair, sorted DESC by sandboxPath
	pathMap  [][2]string // list of (containerPath, sandboxPath) pair, sorted DESC by containerPath
}

func NewContainerRuntime(engine container.Engine, opts ...specdef.Option) (Container, error) {
	spec := new(types.ContainerSpec)
	if err := specdef.Apply(spec, opts...); err != nil {
		return nil, err
	}

	rt := &containerRuntime{
		engine: engine,
		spec:   spec,
	}

	containerPaths := hashset.New[string]()
	sandboxerPaths := hashset.New[string]()
	mountMap := make([][2]string, 0, len(spec.Mounts))
	pathMap := make([][2]string, 0, len(spec.Mounts))
	for _, m := range spec.Mounts {
		cPath, sPath := m.Target, m.Target
		if m.Type == "bind" && m.Source != "" {
			sPath = m.Source
		}
		if s := strings.TrimRight(sPath, "/"); sandboxerPaths.Contains(s) {
			return nil, fmt.Errorf("found duplicate sandbox mount at %q", s)
		} else {
			sandboxerPaths.Add(s)
		}
		if s := strings.TrimRight(cPath, "/"); containerPaths.Contains(s) {
			return nil, fmt.Errorf("found duplicate container mount at %q", s)
		} else {
			containerPaths.Add(s)
		}
		mountMap = append(mountMap, [...]string{sPath, cPath})
		pathMap = append(pathMap, [...]string{cPath, sPath})
	}

	slices.SortFunc(mountMap, func(a, b [2]string) int {
		return strings.Compare(b[0], a[0]) // DESC order by sandboxPath
	})
	slices.SortFunc(pathMap, func(a, b [2]string) int {
		return strings.Compare(b[0], a[0]) // DESC order by containerPath
	})
	rt.mountMap = mountMap
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
	})
}

func (rt *containerRuntime) Build(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (rt *containerRuntime) Run(
	ctx context.Context, image string,
	entrypoint, cmd []string, env map[string]string,
	streams *stream.Streams,
) error {
	// clone env to avoid modify the original
	runEnv := maps.Clone(rt.spec.Environment)
	if runEnv == nil {
		runEnv = make(map[string]string, len(env))
	}
	for k, v := range env {
		if path := MapPath(v, rt.mountMapSeq); path != "" {
			v = path
		}
		runEnv[k] = v
	}

	spec := *rt.spec // clone ContainerSpec
	spec.Image = image
	spec.Entrypoint = entrypoint
	spec.Command = cmd
	spec.Environment = runEnv
	spec.AutoRemove = true

	stdio := new(types.Stdio)
	if streams.Out != nil {
		stdio.Attach |= types.Stdout
	}
	if streams.Err != nil {
		stdio.Attach |= types.Stderr
	}
	_, err := rt.engine.ContainerRun(ctx, &spec, &container.RunOptions{
		Stdio:   stdio,
		Streams: streams,
	})
	return err
}

func (rt *containerRuntime) mountMapSeq(yield func(string, string) bool) {
	for _, pair := range rt.mountMap {
		if !yield(pair[0], pair[1]) {
			return
		}
	}
}
