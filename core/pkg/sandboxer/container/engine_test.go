/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container_test

import (
	"context"
	"testing"

	"drassi.run/core/config"
	mock_container "drassi.run/core/mock/container"
	mock_store "drassi.run/core/mock/store/oci"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	sandboxer_container "drassi.run/core/pkg/sandboxer/container"
	ocistore "drassi.run/core/pkg/store/oci"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestContainerEngineWithProvisioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock_store.NewMockManager(ctrl)
	mockClient := mock_container.NewMockEngine(ctrl)

	img := &ocistore.Image{}
	store.EXPECT().Image(gomock.Any(), "drassi/node:24").Return(img, nil).AnyTimes()
	store.EXPECT().Mount(gomock.Any(), img, gomock.Any()).Return("/var/lib/drassi/node_mount", "layer-node", nil).AnyTimes()
	store.EXPECT().Unmount(gomock.Any(), "layer-node").Return(nil).AnyTimes()

	runtimes := map[string]*config.Runtime{
		"node": {Image: "drassi/node:24"},
	}

	p := provision.New[*types.ContainerSpec](
		runtimes,
		provision.Pull[*types.ContainerSpec](store),
		provision.Mount[*types.ContainerSpec](store, ocistore.WithWritable(true)),
		sandboxer_container.AddBindMount(),
	)

	eng := sandboxer_container.NewWithClient(mockClient, "default:image", p)

	mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, spec *types.ContainerSpec, opts *container.RunOptions) (string, error) {
			require.Len(t, spec.Mounts, 1)
			require.Equal(t, "/var/lib/drassi/node_mount", spec.Mounts[0].Source)
			require.Equal(t, "/opt/drassi/runtimes/node", spec.Mounts[0].Target)
			return "c-123", nil
		},
	)
	mockClient.EXPECT().CopyIn(gomock.Any(), "c-123", gomock.Any()).Return(nil)
	mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-123").Return(&types.ContainerSpec{}, nil)
	mockClient.EXPECT().ContainerRemove(gomock.Any(), gomock.Any()).Return(nil)

	req := &sandboxer.LaunchRequest{}
	resp, err := eng.Launch(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Sandbox)
	require.NotNil(t, resp.JobContainer)
	require.Equal(t, "c-123", resp.JobContainer.Id)

	// Terminate sandbox unmounts layers
	require.NoError(t, resp.Sandbox.Terminate(t.Context()))
}

func TestContainerEngineWithoutProvisioner(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockClient := mock_container.NewMockEngine(ctrl)

	eng := sandboxer_container.NewWithClient(mockClient, "default:image", nil)

	mockClient.EXPECT().ContainerRun(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, spec *types.ContainerSpec, opts *container.RunOptions) (string, error) {
			require.Empty(t, spec.Mounts)
			return "c-456", nil
		},
	)
	mockClient.EXPECT().CopyIn(gomock.Any(), "c-456", gomock.Any()).Return(nil)
	mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-456").Return(&types.ContainerSpec{}, nil)
	mockClient.EXPECT().ContainerRemove(gomock.Any(), gomock.Any()).Return(nil)

	req := &sandboxer.LaunchRequest{}
	resp, err := eng.Launch(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Sandbox)
	require.NotNil(t, resp.JobContainer)
	require.Equal(t, "c-456", resp.JobContainer.Id)

	require.NoError(t, resp.Sandbox.Terminate(t.Context()))
}

func TestContainerFactory(t *testing.T) {
	t.Run("with runtimes and store", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		store := mock_store.NewMockManager(ctrl)

		f := sandboxer_container.NewFactory(sandboxer_container.DefaultConfig())
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
		f := sandboxer_container.NewFactory(sandboxer_container.DefaultConfig())
		f.ProvisionRuntime(map[string]*config.Runtime{
			"node": {Image: "drassi/node:24"},
		})

		_, err := f.Create()
		require.Error(t, err)
		require.Contains(t, err.Error(), "oci store is required")
	})

	t.Run("without runtimes", func(t *testing.T) {
		f := sandboxer_container.NewFactory(sandboxer_container.DefaultConfig())
		eng, err := f.Create()
		require.NoError(t, err)
		require.NotNil(t, eng)
		_ = eng.Close()
	})
}

