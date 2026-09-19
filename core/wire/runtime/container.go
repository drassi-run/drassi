/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_runtime

import (
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/specdef"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	sb "drassi.run/core/pkg/sandboxer/container"
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
	opts := []specdef.Option{
		specdef.AddMount(mounts...),
		specdef.MountApiSocket(engine),
		specdef.SetWorkdir(layout.Workspace()),
	}
	if forge != nil {
		labels := forge.WellKnownLabels()
		labels["run.drassi.runtime"] = "true"

		opts = append(opts, specdef.SetLabels(labels))
	}
	if opt := networkOpt(info); opt != nil {
		opts = append(opts, opt)
	}

	return runtime.NewContainerRuntime(engine, opts...)
}

func networkOpt(info *records.JobInfo) specdef.Option {
	if info.Container != nil {
		if net := info.Container.Network; net != "" {
			return specdef.SetNetwork(net)
		}
	}

	for _, svc := range info.Services {
		if net := svc.Network; net != "" {
			return specdef.SetNetwork(net)
		}
	}

	return nil
}
