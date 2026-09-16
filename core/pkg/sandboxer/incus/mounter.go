/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
)

func newMounter() sandboxer.Mounter {
	return mounter{}
}

type mounter struct{}

func (m mounter) OverlayMounts(target sandboxer.Layout, coarse bool) []*types.Mount {
	if std, ok := target.(sandboxer.StandardLayout); coarse && ok {
		return []*types.Mount{{
			Type:        "bind",
			Source:      string(layout),
			Target:      string(std),
			BindOptions: &types.BindOptions{Recursive: "enabled"},
		}}
	}

	return []*types.Mount{
		{Type: "bind", Source: layout.Workspace(), Target: target.Workspace()},
		{Type: "bind", Source: layout.Temp(), Target: target.Temp()},
		{Type: "bind", Source: layout.Actions(), Target: target.Actions()},
		{Type: "bind", Source: layout.Tools(), Target: target.Tools()},
		{
			Type:        "bind",
			Source:      layout.Runtimes(),
			Target:      target.Runtimes(),
			BindOptions: &types.BindOptions{Recursive: "enabled"},
		},
	}
}

func (m mounter) RuntimeMounts(target sandboxer.Layout) []*types.Mount {
	return []*types.Mount{
		{Type: "bind", Source: layout.Workspace(), Target: target.Workspace()},
		{Type: "bind", Source: layout.Temp(), Target: target.Temp()},
	}
}
