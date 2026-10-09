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
	"github.com/stretchr/testify/require"
)

func TestIncusDiskDevice(t *testing.T) {
	const hostMount = "/var/lib/drassi/storage/overlay/merged"

	t.Run("basic disk device", func(t *testing.T) {
		op := AddDiskDevice("node", hostMount, "/opt/drassi/runtimes/node", config.Runtime{})
		require.Equal(t, "incus/disk-device", op.Name())

		tmpl := new(Template)
		tmpl, err := op.PreLaunch(t.Context(), tmpl)
		require.NoError(t, err)
		require.Contains(t, tmpl.Devices, "runtime-node")
		dev := tmpl.Devices["runtime-node"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, hostMount, dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/node", dev["path"])
		require.Equal(t, "false", dev["readonly"])
	})

	t.Run("readonly disk device", func(t *testing.T) {
		op := AddDiskDevice("node", hostMount, "/opt/drassi/runtimes/node", config.Runtime{ReadOnly: true})
		tmpl := new(Template)
		tmpl, err := op.PreLaunch(t.Context(), tmpl)
		require.NoError(t, err)
		require.Contains(t, tmpl.Devices, "runtime-node")
		dev := tmpl.Devices["runtime-node"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, hostMount, dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/node", dev["path"])
		require.Equal(t, "true", dev["readonly"])
	})

	t.Run("with subpath", func(t *testing.T) {
		op := AddDiskDevice("python", hostMount, "/opt/drassi/runtimes/python", config.Runtime{Subpath: "opt/python"})
		tmpl := new(Template)
		tmpl, err := op.PreLaunch(t.Context(), tmpl)
		require.NoError(t, err)
		require.Contains(t, tmpl.Devices, "runtime-python")
		dev := tmpl.Devices["runtime-python"]
		require.Equal(t, "disk", dev["type"])
		require.Equal(t, filepath.Join(hostMount, "opt/python"), dev["source"])
		require.Equal(t, "/opt/drassi/runtimes/python", dev["path"])
		require.Equal(t, "false", dev["readonly"])
	})
}
