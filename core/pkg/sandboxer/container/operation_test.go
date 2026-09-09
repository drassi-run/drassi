/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"github.com/stretchr/testify/require"
)

func newTestSetup(t *testing.T, name string, cfg *config.Runtime) (*types.ContainerSpec, *provision.Context) {
	t.Helper()
	if cfg == nil {
		cfg = new(config.Runtime)
	}
	spec := new(types.ContainerSpec)
	pctx := provision.NewContext(t.Context(), name, cfg, "/opt/drassi/runtimes/"+name)
	return spec, pctx
}

func TestContainerBindMount(t *testing.T) {
	t.Run("basic bind mount", func(t *testing.T) {
		spec, pctx := newTestSetup(t, "node", nil)
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := AddBindMount(spec)
		require.Equal(t, "container/bind-mount", op.Name())

		require.NoError(t, op.PreLaunch(pctx))
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "bind", spec.Mounts[0].Type)
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/node", spec.Mounts[0].Target)
		require.NoError(t, op.PostLaunch(pctx, nil))
	})

	t.Run("with subpath", func(t *testing.T) {
		spec, pctx := newTestSetup(t, "python", &config.Runtime{Subpath: "opt/python"})
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := AddBindMount(spec)
		require.NoError(t, op.PreLaunch(pctx))
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "bind", spec.Mounts[0].Type)
		require.Equal(t, filepath.Join("/var/lib/drassi/storage/overlay/merged", "opt/python"), spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/python", spec.Mounts[0].Target)
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		spec, pctx := newTestSetup(t, "node", nil)

		op := AddBindMount(spec)
		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}

func TestContainerImageMount(t *testing.T) {
	t.Run("basic image mount", func(t *testing.T) {
		spec, pctx := newTestSetup(t, "node", &config.Runtime{Image: "drassi/node:24"})

		op := AddImageMount(spec)
		require.Equal(t, "container/image-mount", op.Name())

		require.NoError(t, op.PreLaunch(pctx))
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "image", spec.Mounts[0].Type)
		require.Equal(t, "drassi/node:24", spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/node", spec.Mounts[0].Target)
		require.Nil(t, spec.Mounts[0].ImageOptions)
		require.NoError(t, op.PostLaunch(pctx, nil))
	})

	t.Run("with subpath", func(t *testing.T) {
		spec, pctx := newTestSetup(t, "python", &config.Runtime{
			Image:   "drassi/python:3.12",
			Subpath: "opt/python",
		})

		op := AddImageMount(spec)
		require.NoError(t, op.PreLaunch(pctx))
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "image", spec.Mounts[0].Type)
		require.Equal(t, "drassi/python:3.12", spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/python", spec.Mounts[0].Target)
		require.NotNil(t, spec.Mounts[0].ImageOptions)
		require.Equal(t, "opt/python", spec.Mounts[0].ImageOptions.Subpath)
	})
}
