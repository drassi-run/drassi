/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/require"
)

func TestRuntimeConfigParsing(t *testing.T) {
	raw := `
[runtimes.node]
image = "drassi/node:24"
alias = ["node20", "node22"]
executable = "./bin/node"
cmd = ["{0}"]
paths = ["bin"]
`
	var cfg Config[any]
	err := toml.Unmarshal([]byte(raw), &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Runtimes, 1)

	node := cfg.Runtimes["node"]
	require.NotNil(t, node)
	require.Equal(t, "drassi/node:24", node.Image)
	require.Equal(t, []string{"node20", "node22"}, node.Alias)
	require.Equal(t, "./bin/node", node.Executable)
	require.Equal(t, []string{"{0}"}, node.Cmd)
	require.Equal(t, []string{"bin"}, node.Paths)
}

func TestOciStoreConfigParsing(t *testing.T) {
	raw := `
[ocistore]
backend = "containerd"

[ocistore.containerd]
address = "/run/containerd/containerd.sock"
namespace = "default"
fs_snapshotter = "native"
block_snapshotter = "blockfile"
`
	var cfg Config[any]
	err := toml.Unmarshal([]byte(raw), &cfg)
	require.NoError(t, err)
	require.NotNil(t, cfg.OciStore)
	require.Equal(t, "containerd", cfg.OciStore.Backend)
	require.NotNil(t, cfg.OciStore.Containerd)
	require.Equal(t, "/run/containerd/containerd.sock", cfg.OciStore.Containerd.Address)
	require.Equal(t, "default", cfg.OciStore.Containerd.Namespace)
	require.Equal(t, "native", cfg.OciStore.Containerd.FsSnapshotter)
	require.Equal(t, "blockfile", cfg.OciStore.Containerd.BlockSnapshotter)
}

func TestDefaultConfigOciStore(t *testing.T) {
	cfg := DefaultConfig[any]()
	require.NotNil(t, cfg.OciStore)
	require.Equal(t, "containerd", cfg.OciStore.Backend)
	require.NotNil(t, cfg.OciStore.Containerd)
	require.Equal(t, "/run/containerd/containerd.sock", cfg.OciStore.Containerd.Address)
	require.Equal(t, "moby", cfg.OciStore.Containerd.Namespace)
	require.Equal(t, "overlayfs", cfg.OciStore.Containerd.FsSnapshotter)
	require.Equal(t, "", cfg.OciStore.Containerd.BlockSnapshotter)
}
