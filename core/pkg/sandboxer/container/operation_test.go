/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container_test

import (
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer/container"
	"github.com/stretchr/testify/require"
)

func TestContainerBindMount(t *testing.T) {
	t.Run("basic bind mount", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{}, "/opt/drassi/runtimes/node")
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := container.AddBindMount()
		require.Equal(t, "container/bind-mount", op.Name())

		spec := &types.ContainerSpec{}
		spec, err := op.PreLaunch(pctx, spec)
		require.NoError(t, err)
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "bind", spec.Mounts[0].Type)
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/node", spec.Mounts[0].Target)
		require.False(t, spec.Mounts[0].ReadOnly)
	})

	t.Run("readonly bind mount", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{ReadOnly: true}, "/opt/drassi/runtimes/node")
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := container.AddBindMount()
		spec := &types.ContainerSpec{}
		spec, err := op.PreLaunch(pctx, spec)
		require.NoError(t, err)
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "bind", spec.Mounts[0].Type)
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/node", spec.Mounts[0].Target)
		require.True(t, spec.Mounts[0].ReadOnly)
	})

	t.Run("with subpath", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "python", &config.Runtime{Subpath: "opt/python"}, "/opt/drassi/runtimes/python")
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := container.AddBindMount()
		spec := &types.ContainerSpec{}
		spec, err := op.PreLaunch(pctx, spec)
		require.NoError(t, err)
		require.Len(t, spec.Mounts, 1)
		require.Equal(t, "bind", spec.Mounts[0].Type)
		require.Equal(t, filepath.Join("/var/lib/drassi/storage/overlay/merged", "opt/python"), spec.Mounts[0].Source)
		require.Equal(t, "/opt/drassi/runtimes/python", spec.Mounts[0].Target)
		require.False(t, spec.Mounts[0].ReadOnly)
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{}, "/opt/drassi/runtimes/node")
		op := container.AddBindMount()
		spec := &types.ContainerSpec{}
		_, err := op.PreLaunch(pctx, spec)
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}
