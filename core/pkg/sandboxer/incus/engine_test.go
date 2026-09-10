/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus_test

import (
	"context"
	"testing"

	"drassi.run/core/config"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	sandboxer_incus "drassi.run/core/pkg/sandboxer/incus"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIncusProvisionerIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	img := &ocistore.Image{}
	store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).Times(1)
	store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return("/var/lib/drassi/node_mount", "layer-node", nil).Times(1)
	store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).Times(1)

	runtimes := map[string]*config.Runtime{
		"node": {Image: "drassi/node:24"},
	}

	p := provision.New[*sandboxer_incus.Template](
		runtimes,
		provision.Pull[*sandboxer_incus.Template](store),
		provision.Mount[*sandboxer_incus.Template](store, ocistore.WithWritable(true)),
		sandboxer_incus.AddDiskDevice(),
	)

	tmpl := &sandboxer_incus.Template{}
	launched := false
	mockSb := mock_sandboxer.NewMockSandbox(ctrl)
	mockSb.EXPECT().Terminate(gomock.Any()).Return(nil).Times(1)

	launcher := func(ctx context.Context, tmpl *sandboxer_incus.Template) (sandboxer.Sandbox, error) {
		launched = true
		require.Contains(t, tmpl.Devices, "runtime-node")
		require.Equal(t, "disk", tmpl.Devices["runtime-node"]["type"])
		require.Equal(t, "/var/lib/drassi/node_mount", tmpl.Devices["runtime-node"]["source"])
		require.Equal(t, "/opt/drassi/runtimes/node", tmpl.Devices["runtime-node"]["path"])
		return mockSb, nil
	}

	sb, err := p.Launch("/opt/drassi/runtimes", launcher)(t.Context(), tmpl)
	require.NoError(t, err)
	require.True(t, launched)
	require.NotNil(t, sb)

	// Verify unmount cleanup runs on terminate
	require.NoError(t, sb.Terminate(t.Context()))
}

func TestTemplateClone(t *testing.T) {
	t.Run("nil template", func(t *testing.T) {
		var tmpl *sandboxer_incus.Template
		require.Nil(t, tmpl.Clone())
	})

	t.Run("deep copy fields", func(t *testing.T) {
		orig := &sandboxer_incus.Template{
			Name:         "test-instance",
			Image:        "ubuntu:22.04",
			Architecture: "x86_64",
			InstanceSize: "t1.micro",
			Profiles:     []string{"default", "custom"},
			Config: map[string]string{
				"security.nesting": "true",
			},
			Devices: map[string]map[string]string{
				"root": {
					"type": "disk",
					"path": "/",
				},
			},
			Ephemeral: true,
		}

		clone := orig.Clone()
		require.NotNil(t, clone)
		require.Equal(t, orig.Name, clone.Name)
		require.Equal(t, orig.Image, clone.Image)
		require.Equal(t, orig.Architecture, clone.Architecture)
		require.Equal(t, orig.InstanceSize, clone.InstanceSize)
		require.Equal(t, orig.Ephemeral, clone.Ephemeral)
		require.Equal(t, orig.Profiles, clone.Profiles)
		require.Equal(t, orig.Config, clone.Config)
		require.Equal(t, orig.Devices, clone.Devices)

		// Verify Profiles is deep copied
		clone.Profiles[0] = "modified"
		require.Equal(t, "default", orig.Profiles[0])

		// Verify Config is deep copied
		clone.Config["security.nesting"] = "false"
		clone.Config["new.key"] = "val"
		require.Equal(t, "true", orig.Config["security.nesting"])
		require.NotContains(t, orig.Config, "new.key")

		// Verify Devices outer map is deep copied
		clone.Devices["extra"] = map[string]string{"type": "nic"}
		require.NotContains(t, orig.Devices, "extra")

		// Verify Devices inner map is deep copied
		clone.Devices["root"]["path"] = "/mnt"
		require.Equal(t, "/", orig.Devices["root"]["path"])
	})
}

func TestIncusFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	t.Run("with runtimes", func(t *testing.T) {
		f := sandboxer_incus.NewFactory(sandboxer_incus.DefaultConfig())
		f.ProvisionRuntime(store, map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		require.NotPanics(t, func() {
			_, _ = f.Create()
		})
	})

	t.Run("without runtimes", func(t *testing.T) {
		f := sandboxer_incus.NewFactory(sandboxer_incus.DefaultConfig())
		require.NotPanics(t, func() {
			_, _ = f.Create()
		})
	})
}

func TestNew(t *testing.T) {
	t.Run("without panic when nil or omitted", func(t *testing.T) {
		cfg := sandboxer_incus.DefaultConfig()
		require.NotPanics(t, func() {
			_, _ = sandboxer_incus.New(cfg)
		})
		require.NotPanics(t, func() {
			_, _ = sandboxer_incus.New(cfg, nil)
		})
	})
}
