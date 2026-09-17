/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"os"
	"path/filepath"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/sandboxer"
)

func newMounter(jobDir string) sandboxer.Mounter {
	return mounter(jobDir)
}

type mounter string

func (m mounter) OverlayMounts(target sandboxer.Layout, coarse bool) []*types.Mount {
	mounts := make([]*types.Mount, 0)
	layout := sandboxer.StandardLayout(m)

	if std, ok := target.(sandboxer.StandardLayout); coarse && ok {
		mounts = append(mounts, &types.Mount{
			Type:   "bind",
			Source: string(m),
			Target: string(std),
		})
	} else {
		mounts = append(mounts,
			&types.Mount{Type: "bind", Source: layout.Workspace(), Target: target.Workspace()},
			&types.Mount{Type: "bind", Source: layout.Temp(), Target: target.Temp()},
			&types.Mount{Type: "bind", Source: layout.Actions(), Target: target.Actions()},
			&types.Mount{Type: "bind", Source: layout.Tools(), Target: target.Tools()},
		)
	}

	if layout.Runtimes() != "" {
		if entries, err := os.ReadDir(layout.Runtimes()); err == nil {
			for _, e := range entries {
				runtimePath := filepath.Join(layout.Runtimes(), e.Name())
				if runtimePath, err = filepath.EvalSymlinks(runtimePath); err != nil { // follow symlink
					continue
				}
				mounts = append(mounts, &types.Mount{
					Type:     "bind",
					Source:   runtimePath,
					Target:   filepath.Join(target.Runtimes(), e.Name()),
					ReadOnly: true,
				})
			}
		}
	}

	return mounts
}

func (m mounter) RuntimeMounts(target sandboxer.Layout) []*types.Mount {
	layout := sandboxer.StandardLayout(m)
	return []*types.Mount{
		{Type: "bind", Source: layout.Workspace(), Target: target.Workspace()},
		{Type: "bind", Source: layout.Temp(), Target: target.Temp()},
	}
}
