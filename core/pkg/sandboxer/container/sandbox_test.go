/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"testing"

	mock_container "drassi.run/core/mock/container"
	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestExportedHelpers(t *testing.T) {
	layout := DefaultLayout(DefaultJobDir)

	t.Run("DefaultLayout is populated", func(t *testing.T) {
		require.NotEmpty(t, layout.Workspace)
		require.NotEmpty(t, layout.Temp)
		require.NotEmpty(t, layout.Actions)
		require.NotEmpty(t, layout.Tools)
		require.NotEmpty(t, layout.Runtimes)
	})

	t.Run("NewSandbox creates sandbox successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockClient := mock_container.NewMockEngine(ctrl)

		mockClient.EXPECT().CopyIn(gomock.Any(), "c-test-1", gomock.Any()).Return(nil)
		mockClient.EXPECT().ContainerInspect(gomock.Any(), "c-test-1").Return(&types.ContainerSpec{
			Environment: map[string]string{"PATH": "/usr/bin"},
		}, nil)

		sb, err := NewSandbox(context.Background(), mockClient, "c-test-1", layout)
		require.NoError(t, err)
		require.NotNil(t, sb)
		require.Equal(t, layout.Workspace, sb.Layout().Workspace)
	})

	t.Run("Cleanup helper returns cleanup function", func(t *testing.T) {
		called := false
		cu := cleanup(map[string]string{"test": "label"}, func(_ context.Context, opts *container.RemoveOptions) error {
			called = true
			require.Equal(t, "label", opts.Labels["test"])
			return nil
		})
		require.NoError(t, cu(context.Background()))
		require.True(t, called)
	})
}
