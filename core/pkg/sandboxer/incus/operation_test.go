/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus_test

import (
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer/incus"
	"github.com/stretchr/testify/require"
)

func TestIncusDiskDevice(t *testing.T) {
	t.Run("basic disk device", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{}, "/opt/drassi/runtimes/node")
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := incus.AddDiskDevice()
		require.Equal(t, "incus/disk-device", op.Name())

		tmpl := &incus.Template{}
		tmpl, err := op.PreLaunch(pctx, tmpl)
		require.NoError(t, err)
		require.Contains(t, tmpl.Devices, "runtime-node")
		dev := tmpl.Devices["runtime-node"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/node", dev["path"])
	})

	t.Run("with subpath", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "python", &config.Runtime{Subpath: "opt/python"}, "/opt/drassi/runtimes/python")
		pctx.Set(provision.KeyHostMountDir, "/var/lib/drassi/storage/overlay/merged")

		op := incus.AddDiskDevice()
		tmpl := &incus.Template{}
		tmpl, err := op.PreLaunch(pctx, tmpl)
		require.NoError(t, err)
		require.Contains(t, tmpl.Devices, "runtime-python")
		dev := tmpl.Devices["runtime-python"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, filepath.Join("/var/lib/drassi/storage/overlay/merged", "opt/python"), dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/python", dev["path"])
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{}, "/opt/drassi/runtimes/node")
		op := incus.AddDiskDevice()
		tmpl := &incus.Template{}
		_, err := op.PreLaunch(pctx, tmpl)
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}
