/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"testing"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/runtime/provision"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPullOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	op := provision.Pull[any](store)
	require.Equal(t, "pull", op.Name())

	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := provision.NewContext(t.Context(), "node", rtCfg, "/opt/drassi/runtimes/node")

	t.Run("image cached locally", func(t *testing.T) {
		expectedImg := &ocistore.Image{}
		store.EXPECT().Image(pctx, "drassi/node:24").Return(expectedImg, nil)

		cleanup, err := op.Prepare(pctx)
		require.NoError(t, err)
		require.Nil(t, cleanup)

		img, ok := pctx.Get(provision.KeyImage)
		require.True(t, ok)
		require.Equal(t, expectedImg, img)
	})

	t.Run("image not cached - pulls successfully", func(t *testing.T) {
		expectedImg := &ocistore.Image{}
		store.EXPECT().Image(pctx, "drassi/node:24").Return(nil, nil)
		store.EXPECT().Pull(pctx, "drassi/node:24").Return(expectedImg, nil)

		cleanup, err := op.Prepare(pctx)
		require.NoError(t, err)
		require.Nil(t, cleanup)

		img, ok := pctx.Get(provision.KeyImage)
		require.True(t, ok)
		require.Equal(t, expectedImg, img)
	})
}

func TestMountOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	op := provision.Mount[any](store, ocistore.WithWritable(true))
	require.Equal(t, "mount", op.Name())

	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := provision.NewContext(t.Context(), "node", rtCfg, "/opt/drassi/runtimes/node")

	t.Run("missing image returns error", func(t *testing.T) {
		cleanup, err := op.Prepare(pctx)
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "image \"drassi/node:24\" not found in context")
	})

	t.Run("mounts image and returns unmount cleanup", func(t *testing.T) {
		img := &ocistore.Image{}
		pctx.Set(provision.KeyImage, img)

		store.EXPECT().Mount(pctx, img, gomock.Any()).Return("/var/lib/drassi/mount", "layer-123", nil)

		cleanup, err := op.Prepare(pctx)
		require.NoError(t, err)
		require.NotNil(t, cleanup)

		hostDir, ok := pctx.Get(provision.KeyHostMountDir)
		require.True(t, ok)
		require.Equal(t, "/var/lib/drassi/mount", hostDir)

		mountID, ok := pctx.Get(provision.KeyMountID)
		require.True(t, ok)
		require.Equal(t, "layer-123", mountID)

		// Executing cleanup unmounts layer
		store.EXPECT().Unmount(gomock.Any(), "layer-123").Return(nil)
		require.NoError(t, cleanup(t.Context()))
	})

	t.Run("readonly runtime mounts as non-writable", func(t *testing.T) {
		roCfg := &config.Runtime{Image: "drassi/node:24", ReadOnly: true}
		roCtx := provision.NewContext(t.Context(), "node", roCfg, "/opt/drassi/runtimes/node")
		img := &ocistore.Image{}
		roCtx.Set(provision.KeyImage, img)

		store.EXPECT().Mount(roCtx, img, gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *ocistore.Image, opts ...ocistore.MountOption) (string, string, error) {
				require.NotEmpty(t, opts)
				return "/var/lib/drassi/mount-ro", "layer-ro", nil
			},
		)

		cleanup, err := op.Prepare(roCtx)
		require.NoError(t, err)
		require.NotNil(t, cleanup)

		hostDir, ok := roCtx.Get(provision.KeyHostMountDir)
		require.True(t, ok)
		require.Equal(t, "/var/lib/drassi/mount-ro", hostDir)

		mountID, ok := roCtx.Get(provision.KeyMountID)
		require.True(t, ok)
		require.Equal(t, "layer-ro", mountID)

		store.EXPECT().Unmount(gomock.Any(), "layer-ro").Return(nil)
		require.NoError(t, cleanup(t.Context()))
	})
}
