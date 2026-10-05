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

	mock_store "drassi.run/core/mock/store/oci"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPullOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)
	state := new(State)
	const image = "drassi/node:24"

	op := Pull[any](store, state, image)
	require.Equal(t, "pull", op.Name())

	t.Run("pulls successfully", func(t *testing.T) {
		expectedImg := mock_store.NewMockImage(ctrl)
		store.EXPECT().Pull(gomock.Any(), image).Return(expectedImg, nil)

		cleanup, err := op.Prepare(t.Context())
		require.NoError(t, err)
		require.Nil(t, cleanup)

		img, ok := state.Get(KeyImage)
		require.True(t, ok)
		require.Equal(t, expectedImg, img)
	})

	t.Run("pull error returns error", func(t *testing.T) {
		store.EXPECT().Pull(gomock.Any(), image).Return(nil, errors.New("pull failure"))

		cleanup, err := op.Prepare(t.Context())
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "pull image \"drassi/node:24\": pull failure")
	})
}

func TestMountOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)
	state := new(State)

	op := Mount[any](store, state, ocistore.WithWritable(true))
	require.Equal(t, "mount", op.Name())

	t.Run("missing image returns error", func(t *testing.T) {
		cleanup, err := op.Prepare(t.Context())
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "image not found in state")
	})

	t.Run("mounts image and returns unmount cleanup", func(t *testing.T) {
		var img ocistore.Image = mock_store.NewMockImage(ctrl)
		state.Set(KeyImage, img)

		released := false
		mnt := &ocistore.Mount{
			Target: "/var/lib/drassi/mount",
			Release: func(ctx context.Context) error {
				released = true
				return nil
			},
		}

		store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return(mnt, nil)

		cleanup, err := op.Prepare(t.Context())
		require.NoError(t, err)
		require.NotNil(t, cleanup)

		require.NoError(t, cleanup(t.Context()))
		require.True(t, released)
	})

	t.Run("mount error returns error", func(t *testing.T) {
		var img ocistore.Image = mock_store.NewMockImage(ctrl)
		state.Set(KeyImage, img)

		store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return(nil, errors.New("mount failure"))

		cleanup, err := op.Prepare(t.Context())
		require.Error(t, err)
		require.Nil(t, cleanup)
		require.Contains(t, err.Error(), "mount image: mount failure")
	})
}
