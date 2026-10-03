/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSandboxerCustomTOML(t *testing.T) {
	t.Run("UnmarshalTOML with inlined config", func(t *testing.T) {
		raw := `
[sandboxer]
provider = "docker"
endpoint = "unix:///var/run/docker.sock"

[sandboxer.template]
image = "ghcr.io/drassi-run/ubuntu:26.04"
privileged = true
`
		var cfg Config[any]
		err := Unmarshal([]byte(raw), &cfg)
		require.NoError(t, err)
		require.NotNil(t, cfg.Sandboxer)
		require.Equal(t, "docker", cfg.Sandboxer.Provider)
		require.JSONEq(t, `{"endpoint":"unix:///var/run/docker.sock","template":{"image":"ghcr.io/drassi-run/ubuntu:26.04","privileged":true}}`, string(cfg.Sandboxer.Config))
	})

	t.Run("MarshalTOML round trip", func(t *testing.T) {
		sb := &Sandboxer{
			Provider: "docker",
			Config:   []byte(`{"endpoint":"unix:///var/run/docker.sock","template":{"image":"ghcr.io/drassi-run/ubuntu:26.04"}}`),
		}
		data, err := sb.MarshalTOML()
		require.NoError(t, err)

		var parsed Sandboxer
		err = parsed.UnmarshalTOML(data)
		require.NoError(t, err)
		require.Equal(t, "docker", parsed.Provider)
		require.JSONEq(t, string(sb.Config), string(parsed.Config))
	})

	t.Run("MarshalTOML without config", func(t *testing.T) {
		sb := &Sandboxer{
			Provider: "host",
		}
		data, err := sb.MarshalTOML()
		require.NoError(t, err)

		var parsed Sandboxer
		err = parsed.UnmarshalTOML(data)
		require.NoError(t, err)
		require.Equal(t, "host", parsed.Provider)
		require.Nil(t, parsed.Config)
	})
}
