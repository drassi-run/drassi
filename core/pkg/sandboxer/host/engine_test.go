/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host_test

import (
	"os"
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	sandboxer_host "drassi.run/core/pkg/sandboxer/host"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHostEngineWithProvisioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	tempDir := t.TempDir()
	mountDir := filepath.Join(tempDir, "node_mount")
	require.NoError(t, os.MkdirAll(mountDir, 0755))

	img := &ocistore.Image{}
	store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).Times(1)
	store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return(mountDir, "layer-node", nil).Times(1)
	store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).Times(1)

	runtimeDir := filepath.Join(tempDir, "opt_drassi_runtimes")
	cfg := &sandboxer_host.Config{
		RootDir:    tempDir,
		RuntimeDir: runtimeDir,
	}

	runtimes := map[string]*config.Runtime{
		"node": {Image: "drassi/node:24"},
	}

	p := provision.New[*sandboxer.LaunchRequest](
		runtimes,
		func(name string) string { return filepath.Join(runtimeDir, name) },
		provision.Pull[*sandboxer.LaunchRequest](store),
		provision.Mount[*sandboxer.LaunchRequest](store, ocistore.WithWritable(true)),
		sandboxer_host.Symlink[*sandboxer.LaunchRequest](),
	)

	eng, err := sandboxer_host.New(cfg, p)
	require.NoError(t, err)

	req := &sandboxer.LaunchRequest{
		Forge: &records.Forge{
			Repository: "drassi/test",
			Workflow:   "build.yml",
			Job:        "test",
			RunId:      "1",
			RunAttempt: "1",
		},
	}

	resp, err := eng.Launch(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Sandbox)

	// Symlink is created
	symlinkPath := filepath.Join(runtimeDir, "node")
	target, err := os.Readlink(symlinkPath)
	require.NoError(t, err)
	require.Equal(t, mountDir, target)

	// Terminate sandbox unmounts layers and removes workspace
	require.NoError(t, resp.Sandbox.Terminate(t.Context()))
}

func TestHostEngineWithoutProvisioner(t *testing.T) {
	tempDir := t.TempDir()
	runtimeDir := filepath.Join(tempDir, "opt_drassi_runtimes")
	cfg := &sandboxer_host.Config{
		RootDir:    tempDir,
		RuntimeDir: runtimeDir,
	}

	t.Run("omitted provisioner", func(t *testing.T) {
		eng, err := sandboxer_host.New(cfg)
		require.NoError(t, err)

		req := &sandboxer.LaunchRequest{
			Forge: &records.Forge{
				Repository: "drassi/test",
				Workflow:   "build.yml",
				Job:        "test",
				RunId:      "1",
				RunAttempt: "1",
			},
		}

		resp, err := eng.Launch(t.Context(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Sandbox)

		require.NoError(t, resp.Sandbox.Terminate(t.Context()))
	})

	t.Run("nil provisioner", func(t *testing.T) {
		eng, err := sandboxer_host.New(cfg, nil)
		require.NoError(t, err)

		req := &sandboxer.LaunchRequest{
			Forge: &records.Forge{
				Repository: "drassi/test",
				Workflow:   "build.yml",
				Job:        "test",
				RunId:      "1",
				RunAttempt: "1",
			},
		}

		resp, err := eng.Launch(t.Context(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Sandbox)

		require.NoError(t, resp.Sandbox.Terminate(t.Context()))
	})
}

func TestHostFactory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	t.Run("with runtimes", func(t *testing.T) {
		cfg := sandboxer_host.DefaultConfig()
		cfg.RootDir = t.TempDir()
		cfg.RuntimeDir = filepath.Join(cfg.RootDir, "runtimes")
		f := sandboxer_host.NewFactory(cfg)
		f.ProvisionRuntime(store, map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})

	t.Run("without runtimes", func(t *testing.T) {
		cfg := sandboxer_host.DefaultConfig()
		cfg.RootDir = t.TempDir()
		cfg.RuntimeDir = filepath.Join(cfg.RootDir, "runtimes")
		f := sandboxer_host.NewFactory(cfg)
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})
}

func TestNew(t *testing.T) {
	t.Run("without panic when nil or omitted", func(t *testing.T) {
		cfg := sandboxer_host.DefaultConfig()
		cfg.RootDir = t.TempDir()
		eng1, err := sandboxer_host.New(cfg)
		require.NoError(t, err)
		require.NotNil(t, eng1)
		_ = eng1.Close()

		eng2, err := sandboxer_host.New(cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, eng2)
		_ = eng2.Close()
	})
}
