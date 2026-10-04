/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision

import (
	"context"
	"errors"
	"testing"

	"drassi.run/core/config"
	mock_store "drassi.run/core/mock/store/oci"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPullOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	op := Pull[any](store)
	require.Equal(t, "pull", op.Name())

	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := NewContext(t.Context(), "node", rtCfg, "/opt/drassi/runtimes/node")

	t.Run("nil store panics", func(t *testing.T) {
		require.Panics(t, func() {
			Pull[any](nil)
		})
	})

	t.Run("pulls successfully", func(t *testing.T) {
		expectedImg := mock_store.NewMockImage(ctrl)
		store.EXPECT().Pull(pctx, "drassi/node:24").Return(expectedImg, nil)

		cleanup, err := op.Prepare(pctx)
		require.NoError(t, err)
		require.Nil(t, cleanup)

		img, ok := pctx.Get(KeyImage)
		require.True(t, ok)
		require.Equal(t, expectedImg, img)
	})

	t.Run("pull error returns error", func(t *testing.T) {
		store.EXPECT().Pull(pctx, "drassi/node:24").Return(nil, errors.New("pull failure"))

		cleanup, err := op.Prepare(pctx)
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "pull image \"drassi/node:24\": pull failure")
	})
}

func TestMountOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)

	op := Mount[any](store, ocistore.WithWritable(true))
	require.Equal(t, "mount", op.Name())

	rtCfg := &config.Runtime{Image: "drassi/node:24"}
	pctx := NewContext(t.Context(), "node", rtCfg, "/opt/drassi/runtimes/node")

	t.Run("nil store panics", func(t *testing.T) {
		require.Panics(t, func() {
			Mount[any](nil)
		})
	})

	t.Run("missing image returns error", func(t *testing.T) {
		cleanup, err := op.Prepare(pctx)
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "image \"drassi/node:24\" not found in context")
	})

	t.Run("mounts image and returns unmount cleanup", func(t *testing.T) {
		var img ocistore.Image = mock_store.NewMockImage(ctrl)
		pctx.Set(KeyImage, img)

		released := false
		mnt := &ocistore.Mount{
			Target: "/var/lib/drassi/mount",
			Release: func(ctx context.Context) error {
				released = true
				return nil
			},
		}

		store.EXPECT().Mount(pctx, img, gomock.Any()).Return(mnt, nil)

		cleanup, err := op.Prepare(pctx)
		require.NoError(t, err)
		require.NotNil(t, cleanup)

		hostDir, ok := pctx.Get(KeyHostMountDir)
		require.True(t, ok)
		require.Equal(t, "/var/lib/drassi/mount", hostDir)

		require.NoError(t, cleanup(t.Context()))
		require.True(t, released)
	})

	t.Run("mount error returns error", func(t *testing.T) {
		var img ocistore.Image = mock_store.NewMockImage(ctrl)
		pctx.Set(KeyImage, img)

		store.EXPECT().Mount(pctx, img, gomock.Any()).Return(nil, errors.New("mount failure"))

		cleanup, err := op.Prepare(pctx)
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "mount image \"drassi/node:24\": mount failure")
	})

	t.Run("readonly runtime mounts as non-writable", func(t *testing.T) {
		roCfg := &config.Runtime{Image: "drassi/node:24", ReadOnly: true}
		roCtx := NewContext(t.Context(), "node", roCfg, "/opt/drassi/runtimes/node")
		var img ocistore.Image = mock_store.NewMockImage(ctrl)
		roCtx.Set(KeyImage, img)

		released := false
		mnt := &ocistore.Mount{
			Target: "/var/lib/drassi/mount-ro",
			Release: func(ctx context.Context) error {
				released = true
				return nil
			},
		}

		store.EXPECT().Mount(roCtx, img, gomock.Any()).Return(mnt, nil)

		cleanup, err := op.Prepare(roCtx)
		require.NoError(t, err)
		require.NotNil(t, cleanup)

		hostDir, ok := roCtx.Get(KeyHostMountDir)
		require.True(t, ok)
		require.Equal(t, "/var/lib/drassi/mount-ro", hostDir)

		require.NoError(t, cleanup(t.Context()))
		require.True(t, released)
	})
}
