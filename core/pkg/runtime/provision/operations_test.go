/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package provision_test

import (
	"context"
	"errors"
	"testing"

	"drassi.run/core/config"
	mock_ocistore "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestSetup(t *testing.T, name, image string) (*mock_ocistore.MockManager, *provision.Context) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mgr := mock_ocistore.NewMockManager(ctrl)
	cfg := &config.Runtime{
		Image: image,
		Paths: []string{"bin"},
	}
	pctx := provision.NewContext(context.Background(), name, cfg, "/opt/drassi/runtimes/"+name)
	return mgr, pctx
}

func TestSharedOperations(t *testing.T) {
	mgr, pctx := newTestSetup(t, "node", "drassi/node:24")
	img := &ocistore.Image{ID: "img-node", Names: []string{"drassi/node:24"}}

	mgr.EXPECT().Image(pctx, "drassi/node:24").Return(nil, nil)
	mgr.EXPECT().Pull(pctx, "drassi/node:24").Return(img, nil)
	mgr.EXPECT().Mount(pctx, img, gomock.Any()).Return("/var/lib/drassi/storage/overlay/merged", "layer-123", nil)

	p := provision.NewPipeline(
		provision.Pull(mgr),
		provision.Mount(mgr, ocistore.WithWritable(true)),
	)

	err := p.PreLaunch(pctx)
	require.NoError(t, err)
	require.Equal(t, "/var/lib/drassi/storage/overlay/merged", pctx.MustGet(provision.KeyHostMountDir))
	require.Equal(t, "layer-123", pctx.MustGet(provision.KeyMountID))
	require.Equal(t, img, pctx.MustGet(provision.KeyImage))

	err = p.PostLaunch(pctx, nil)
	require.NoError(t, err)
}

func TestPullOperation(t *testing.T) {
	t.Run("operation metadata", func(t *testing.T) {
		mgr, _ := newTestSetup(t, "python", "python:3.12")
		op := provision.Pull(mgr)
		require.Equal(t, "pull", op.Name())
		require.NoError(t, op.PostLaunch(nil, nil))
	})

	t.Run("image already exists - pull skipped", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "python", "python:3.12")
		img := &ocistore.Image{ID: "img-existing", Names: []string{"python:3.12"}}

		mgr.EXPECT().Image(pctx, "python:3.12").Return(img, nil)

		op := provision.Pull(mgr)
		err := op.PreLaunch(pctx)
		require.NoError(t, err)
		require.Equal(t, img, pctx.MustGet(provision.KeyImage))
	})

	t.Run("image does not exist - pull triggered", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "python", "python:3.12")
		img := &ocistore.Image{ID: "img-new", Names: []string{"python:3.12"}}

		mgr.EXPECT().Image(pctx, "python:3.12").Return(nil, nil)
		mgr.EXPECT().Pull(pctx, "python:3.12").Return(img, nil)

		op := provision.Pull(mgr)
		err := op.PreLaunch(pctx)
		require.NoError(t, err)
		require.Equal(t, img, pctx.MustGet(provision.KeyImage))
	})

	t.Run("image check error", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "python", "python:3.12")

		mgr.EXPECT().Image(pctx, "python:3.12").Return(nil, errors.New("storage backend disconnected"))

		op := provision.Pull(mgr)
		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `check image "python:3.12": storage backend disconnected`)
	})

	t.Run("pull error", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "python", "python:3.12")

		mgr.EXPECT().Image(pctx, "python:3.12").Return(nil, nil)
		mgr.EXPECT().Pull(pctx, "python:3.12").Return(nil, errors.New("network timeout"))

		op := provision.Pull(mgr)
		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `pull image "python:3.12": network timeout`)
	})
}

func TestMountOperation(t *testing.T) {
	t.Run("operation metadata", func(t *testing.T) {
		mgr, _ := newTestSetup(t, "node", "node:24")
		op := provision.Mount(mgr)
		require.Equal(t, "mount", op.Name())
		require.NoError(t, op.PostLaunch(nil, nil))
	})

	t.Run("mount success with options", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "node", "node:24")
		img := &ocistore.Image{ID: "img-node", Names: []string{"node:24"}}
		pctx.Set(provision.KeyImage, img)

		mgr.EXPECT().Mount(pctx, img, gomock.Any()).DoAndReturn(func(_ context.Context, _ *ocistore.Image, opts ...ocistore.MountOption) (string, string, error) {
			require.Len(t, opts, 1)
			return "/var/lib/drassi/storage/overlay/merged", "layer-123", nil
		})

		op := provision.Mount(mgr, ocistore.WithWritable(true))
		err := op.PreLaunch(pctx)
		require.NoError(t, err)
		require.Equal(t, "/var/lib/drassi/storage/overlay/merged", pctx.MustGet(provision.KeyHostMountDir))
		require.Equal(t, "layer-123", pctx.MustGet(provision.KeyMountID))
		require.Equal(t, img, pctx.MustGet(provision.KeyImage))
	})

	t.Run("image not found in context error", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "node", "node:24")
		op := provision.Mount(mgr)

		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `image "node:24" not found in context: pull operation must be used first`)
	})

	t.Run("mount error", func(t *testing.T) {
		mgr, pctx := newTestSetup(t, "node", "node:24")
		img := &ocistore.Image{ID: "img-node", Names: []string{"node:24"}}
		pctx.Set(provision.KeyImage, img)

		mgr.EXPECT().Mount(pctx, img, gomock.Any()).Return("", "", errors.New("overlay mount failed"))

		op := provision.Mount(mgr)
		err := op.PreLaunch(pctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), `mount image "node:24": overlay mount failed`)

		_, ok := pctx.Get(provision.KeyHostMountDir)
		require.False(t, ok)
		_, ok = pctx.Get(provision.KeyMountID)
		require.False(t, ok)
	})
}

func TestOpFunc(t *testing.T) {
	t.Run("metadata and default execution", func(t *testing.T) {
		op := &provision.OpFunc{}
		require.Equal(t, "func", op.Name())
		require.NoError(t, op.PreLaunch(nil))
		require.NoError(t, op.PostLaunch(nil, nil))
	})

	t.Run("custom callbacks", func(t *testing.T) {
		preCalled, postCalled := false, false
		op := &provision.OpFunc{
			PreFunc: func(_ *provision.Context) error {
				preCalled = true
				return nil
			},
			PostFunc: func(_ *provision.Context, _ sandboxer.Sandbox) error {
				postCalled = true
				return nil
			},
		}

		require.NoError(t, op.PreLaunch(nil))
		require.True(t, preCalled)

		require.NoError(t, op.PostLaunch(nil, nil))
		require.True(t, postCalled)
	})
}
