/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"testing"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Template)
	require.Equal(t, container.DefaultImage, cfg.Template.Image)
	require.Empty(t, cfg.Endpoint)
}

func TestFactoryRegistration(t *testing.T) {
	sbConfig := &config.Sandboxer{
		Provider: config.ProviderDocker,
	}
	f, err := sandboxer.NewFactory(sbConfig)
	require.NoError(t, err)
	require.NotNil(t, f)
}

func TestFactoryWithTemplateConfig(t *testing.T) {
	rawToml := `
endpoint = "unix:///var/run/docker.sock"
[template]
image = "ghcr.io/drassi-run/ubuntu:26.04"
network_mode = "host"
privileged = true
user = "1000:1000"
volumes = [
    "/var/run/docker.sock:/var/run/docker.sock",
    "/data:/data:ro"
]
ports = [
    "8080:80"
]
cap_add = ["SYS_ADMIN"]
devices = ["/dev/kvm"]
cpus = "2"
memory = 2147483648
`
	sbConfig := &config.Sandboxer{
		Provider: config.ProviderDocker,
		Config:   []byte(rawToml),
	}
	f, err := sandboxer.NewFactory(sbConfig)
	require.NoError(t, err)
	require.NotNil(t, f)

	fact, ok := f.(*factory)
	require.True(t, ok)
	require.Equal(t, "unix:///var/run/docker.sock", fact.cfg.Endpoint)
	require.NotNil(t, fact.cfg.Template)
	tmpl := fact.cfg.Template
	require.Equal(t, "ghcr.io/drassi-run/ubuntu:26.04", tmpl.Image)
	require.Equal(t, "host", tmpl.NetworkMode)
	require.True(t, tmpl.Privileged)
	require.Equal(t, "1000:1000", tmpl.User)
	require.Len(t, tmpl.Mounts, 2)
	require.Equal(t, "/var/run/docker.sock", tmpl.Mounts[0].Source)
	require.Equal(t, "/var/run/docker.sock", tmpl.Mounts[0].Target)
	require.False(t, tmpl.Mounts[0].ReadOnly)
	require.Equal(t, "/data", tmpl.Mounts[1].Source)
	require.Equal(t, "/data", tmpl.Mounts[1].Target)
	require.True(t, tmpl.Mounts[1].ReadOnly)
	require.Len(t, tmpl.Publish, 1)
	require.Equal(t, uint16(8080), tmpl.Publish[0].HostPort)
	require.Equal(t, uint16(80), tmpl.Publish[0].ContainerPort)
	require.Equal(t, []string{"SYS_ADMIN"}, tmpl.CapAdd)
	require.Equal(t, []string{"/dev/kvm"}, tmpl.Devices)
	require.Equal(t, "2", tmpl.CPUS)
	require.Equal(t, int64(2147483648), tmpl.Memory)
}

func TestFactoryCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	t.Run("without runtimes creates engine successfully", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})

	t.Run("with runtimes and store creates engine successfully", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		f.SetOciStore(store)
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})

	t.Run("with runtimes but missing store returns error", func(t *testing.T) {
		f := NewFactory(DefaultConfig())
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		_, err := f.Create()
		require.Error(t, err)
		require.Contains(t, err.Error(), "oci store is required")
	})
}

func TestNew(t *testing.T) {
	t.Run("with nil config uses DefaultConfig without panicking", func(t *testing.T) {
		eng, err := New(nil, nil)
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})
}
