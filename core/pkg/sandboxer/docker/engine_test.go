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
	require.Equal(t, container.DefaultImage, cfg.Image)
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
