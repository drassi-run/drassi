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
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/host"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHostSymlink(t *testing.T) {
	ctrl := gomock.NewController(t)
	runtimesDir := filepath.Join(t.TempDir(), "runtimes")
	require.NoError(t, os.MkdirAll(runtimesDir, 0755))

	op := host.Symlink[any]()
	require.Equal(t, "host/symlink", op.Name())

	sb := mock_sandboxer.NewMockSandbox(ctrl)
	sb.EXPECT().Layout().Return(&sandboxer.Layout{Runtimes: runtimesDir}).AnyTimes()

	t.Run("basic symlink", func(t *testing.T) {
		mountDir := filepath.Join(t.TempDir(), "mount_merged")
		require.NoError(t, os.MkdirAll(mountDir, 0755))

		pctx := provision.NewContext(t.Context(), "node", &config.Runtime{}, filepath.Join(runtimesDir, "node"))
		pctx.Set(provision.KeyHostMountDir, mountDir)

		resSb, err := op.PostLaunch(pctx, sb)
		require.NoError(t, err)
		require.Equal(t, sb, resSb)

		target := filepath.Join(runtimesDir, "node")
		targetInfo, err := os.Lstat(target)
		require.NoError(t, err)
		require.True(t, targetInfo.Mode()&os.ModeSymlink != 0)
	})

	t.Run("with subpath", func(t *testing.T) {
		subMountDir := filepath.Join(t.TempDir(), "sub_merged")
		require.NoError(t, os.MkdirAll(filepath.Join(subMountDir, "custom/sub"), 0755))

		pctx := provision.NewContext(t.Context(), "custom", &config.Runtime{Subpath: "custom/sub"}, filepath.Join(runtimesDir, "custom"))
		pctx.Set(provision.KeyHostMountDir, subMountDir)

		_, err := op.PostLaunch(pctx, sb)
		require.NoError(t, err)

		targetSub := filepath.Join(runtimesDir, "custom")
		linkTarget, err := os.Readlink(targetSub)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(subMountDir, "custom/sub"), linkTarget)
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		pctx := provision.NewContext(t.Context(), "missing", &config.Runtime{}, filepath.Join(runtimesDir, "missing"))
		_, err := op.PostLaunch(pctx, sb)
		require.Error(t, err)
		require.Contains(t, err.Error(), "host mount directory not set in context")
	})
}
