/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
)

func newMounter(volume string) sandboxer.Mounter {
	return mounter(volume)
}

type mounter string

func (m mounter) OverlayMounts(_ sandboxer.Layout, _ bool) []*types.Mount {
	// docker does NOT support overlay
	return nil
}

func (m mounter) RuntimeMounts(target sandboxer.Layout) []*types.Mount {
	return []*types.Mount{
		m.subMount(target.Workspace(), "workspace"),
		m.subMount(target.Temp(), "temp"),
	}
}

func (m mounter) subMount(target string, subpath string) *types.Mount {
	return &types.Mount{
		Type:          "volume",
		Source:        string(m),
		Target:        target,
		VolumeOptions: &types.VolumeOptions{SubPath: subpath},
	}
}
