/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config_test

import (
	"testing"

	"drassi.run/core/config"
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
	var cfg config.Config[any]
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
