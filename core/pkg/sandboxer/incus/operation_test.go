/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"github.com/stretchr/testify/require"
)

func newTestSetup(t *testing.T, name string, cfg *config.Runtime) (*Template, *provision.Context) {
	t.Helper()
	if cfg == nil {
		cfg = new(config.Runtime)
	}
	tmpl := new(Template)
	pctx := provision.NewContext(t.Context(), name, cfg, "/opt/drassi/runtimes/"+name)
	return tmpl, pctx
}

func TestIncusDiskDevice(t *testing.T) {
	t.Run("basic disk device", func(t *testing.T) {
		tmpl, pctx := newTestSetup(t, "node", nil)
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := AddDiskDevice(tmpl)
		require.Equal(t, "incus/disk-device", op.Name())

		require.NoError(t, op.PreLaunch(pctx))
		require.Contains(t, tmpl.Devices, "runtime-node")
		dev := tmpl.Devices["runtime-node"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/node", dev["path"])
		require.NoError(t, op.PostLaunch(pctx, nil))
	})

	t.Run("with subpath", func(t *testing.T) {
		tmpl, pctx := newTestSetup(t, "python", &config.Runtime{Subpath: "opt/python"})
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := AddDiskDevice(tmpl)
		require.NoError(t, op.PreLaunch(pctx))
		require.Contains(t, tmpl.Devices, "runtime-python")
		dev := tmpl.Devices["runtime-python"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, filepath.Join("/var/lib/drassi/storage/overlay/merged", "opt/python"), dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/python", dev["path"])
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		tmpl, pctx := newTestSetup(t, "node", nil)

		op := AddDiskDevice(tmpl)
		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}
