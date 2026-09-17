/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_runtime

import (
	"slices"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	sb "drassi.run/core/pkg/sandboxer/container"
	. "drassi.run/core/util/types"
)

func NewContainerRuntime(
	engine container.Engine,
	sandbox sandboxer.Sandbox,
	mounter sandboxer.Mounter,
	info *records.JobInfo,
	forge *records.Forge,
) (runtime.Container, error) {
	layout := sandboxer.Layout(sb.DefaultLayout)
	if info.Container != nil { // if running in job container, use the same layout
		layout = sandbox.Layout()
	}

	mounts := mounter.RuntimeMounts(layout)
	mountPairs := make([]Pair[string, *types.Mount], len(mounts))
	for i, m := range mounts {
		mountPairs[i] = Pair[string, *types.Mount]{
			Key: m.Target, Value: m,
		}
	}

	opts := []runtime.ContainerRuntimeOption{
		runtime.WithMounts(mountPairs),
		staticMountOpt(engine.Address(), mounts),
		runtime.WithWorkDir(layout.Workspace()),
	}
	if opt := labelsOpt(forge); opt != nil {
		opts = append(opts, opt)
	}
	if opt := networkOpt(info); opt != nil {
		opts = append(opts, opt)
	}

	return runtime.NewContainerRuntime(engine, opts...)
}

func labelsOpt(forge *records.Forge) runtime.ContainerRuntimeOption {
	labels := forge.WellKnownLabels()
	return runtime.WithLabels(labels)
}

func networkOpt(info *records.JobInfo) runtime.ContainerRuntimeOption {
	if info.Container != nil {
		if net := info.Container.Network; net != "" {
			return runtime.WithNetwork(net)
		}
	}

	for _, svc := range info.Services {
		if net := svc.Network; net != "" {
			return runtime.WithNetwork(net)
		}
	}

	return nil
}

func staticMountOpt(path string, sbMounts []*types.Mount) runtime.ContainerRuntimeOption {
	path = strings.TrimPrefix(path, "unix://")

	if sbMounts == nil {
		mount := &types.Mount{
			Type:   "bind",
			Source: path,
			Target: path,
		}
		mounts := []Pair[string, *types.Mount]{
			{Key: path, Value: mount},
		}
		return runtime.WithMounts(mounts)
	}

	slices.SortFunc(sbMounts, func(a, b *types.Mount) int {
		return strings.Compare(b.Source, a.Source) // DESC order
	})
	seq := func(yield func(string, string) bool) {
		for _, m := range sbMounts {
			if !yield(m.Source, m.Target) {
				return
			}
		}
	}
	sandboxPath := runtime.MapPath(path, seq)
	mount := &types.Mount{
		Type:   "bind",
		Source: path,
		Target: path,
	}
	mounts := []Pair[string, *types.Mount]{
		{Key: sandboxPath, Value: mount},
	}
	return runtime.WithMounts(mounts)
}
