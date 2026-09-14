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
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/container/types"
	. "drassi.run/core/pkg/runtime/provision"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func runBindMount(t *testing.T, name string, rt *config.Runtime, targetDir, hostMountDir string) (*types.Mount, error) {
	t.Helper()
	pctx := NewContext(t.Context(), name, rt, targetDir)
	if hostMountDir != "" {
		pctx.Set(KeyHostMountDir, hostMountDir)
	}

	op := AddBindMount()
	require.Equal(t, "container/bind-mount", op.Name())

	spec, err := op.PreLaunch(pctx, new(types.ContainerSpec))
	if err != nil {
		return nil, err
	}

	require.Len(t, spec.Mounts, 1)
	require.Equal(t, "bind", spec.Mounts[0].Type)
	return spec.Mounts[0], nil
}

func TestContainerBindMount(t *testing.T) {
	const hostMount = "/var/lib/drassi/storage/overlay/merged"

	t.Run("basic bind mount", func(t *testing.T) {
		mount, err := runBindMount(t, "node", &config.Runtime{}, "/opt/drassi/runtimes/node", hostMount)
		require.NoError(t, err)
		require.Equal(t, hostMount, mount.Source)
		require.Equal(t, "/opt/drassi/runtimes/node", mount.Target)
		require.False(t, mount.ReadOnly)
	})

	t.Run("readonly bind mount", func(t *testing.T) {
		mount, err := runBindMount(t, "node", &config.Runtime{ReadOnly: true}, "/opt/drassi/runtimes/node", hostMount)
		require.NoError(t, err)
		require.Equal(t, hostMount, mount.Source)
		require.Equal(t, "/opt/drassi/runtimes/node", mount.Target)
		require.True(t, mount.ReadOnly)
	})

	t.Run("with subpath", func(t *testing.T) {
		mount, err := runBindMount(t, "python", &config.Runtime{Subpath: "opt/python"}, "/opt/drassi/runtimes/python", hostMount)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(hostMount, "opt/python"), mount.Source)
		require.Equal(t, "/opt/drassi/runtimes/python", mount.Target)
		require.False(t, mount.ReadOnly)
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		_, err := runBindMount(t, "node", &config.Runtime{}, "/opt/drassi/runtimes/node", "")
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}

func TestNewProvisioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	t.Run("with runtimes and store", func(t *testing.T) {
		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}
		p := NewProvisioner(store, runtimes)
		require.NotNil(t, p)
	})

	t.Run("with runtimes but missing store panics", func(t *testing.T) {
		runtimes := map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		}
		require.Panics(t, func() {
			_ = NewProvisioner(nil, runtimes)
		})
	})

	t.Run("without runtimes returns nil provisioner", func(t *testing.T) {
		p := NewProvisioner(store, nil)
		require.Nil(t, p)
	})
}
