/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"os"
	"path/filepath"
	"testing"

	"drassi.run/core/config"
	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestSetup(t *testing.T, name string, cfg *config.Runtime) (provision.Operation, *provision.Context, *mock_sandboxer.MockSandbox, string) {
	t.Helper()
	if cfg == nil {
		cfg = new(config.Runtime)
	}
	runtimesDir := filepath.Join(t.TempDir(), "runtimes")
	require.NoError(t, os.MkdirAll(runtimesDir, 0755))

	pctx := provision.NewContext(t.Context(), name, cfg, filepath.Join(runtimesDir, name))
	op := Symlink()
	sb := mock_sandboxer.NewMockSandbox(gomock.NewController(t))
	sb.EXPECT().Layout().Return(&sandboxer.Layout{Runtimes: runtimesDir}).AnyTimes()
	return op, pctx, sb, runtimesDir
}

func TestHostSymlink(t *testing.T) {
	t.Run("basic symlink", func(t *testing.T) {
		op, pctx, sb, runtimesDir := newTestSetup(t, "node", nil)
		require.Equal(t, "host/symlink", op.Name())
		require.NoError(t, op.PreLaunch(pctx))

		mountDir := filepath.Join(t.TempDir(), "mount_merged")
		require.NoError(t, os.MkdirAll(mountDir, 0755))
		pctx.Set(provision.KeyHostMountDir, mountDir)

		require.NoError(t, op.PostLaunch(pctx, sb))

		target := filepath.Join(runtimesDir, "node")
		targetInfo, err := os.Lstat(target)
		require.NoError(t, err)
		require.True(t, targetInfo.Mode()&os.ModeSymlink != 0)
	})

	t.Run("with subpath", func(t *testing.T) {
		op, pctx, sb, runtimesDir := newTestSetup(t, "custom", &config.Runtime{Subpath: "custom/sub"})

		subMountDir := filepath.Join(t.TempDir(), "sub_merged")
		require.NoError(t, os.MkdirAll(filepath.Join(subMountDir, "custom/sub"), 0755))
		pctx.Set(provision.KeyHostMountDir, subMountDir)

		require.NoError(t, op.PostLaunch(pctx, sb))

		targetSub := filepath.Join(runtimesDir, "custom")
		linkTarget, err := os.Readlink(targetSub)
		require.NoError(t, err)
		require.Equal(t, filepath.Join(subMountDir, "custom/sub"), linkTarget)
	})

	t.Run("missing host mount dir", func(t *testing.T) {
		op, pctx, sb, _ := newTestSetup(t, "missing", nil)

		err := op.PostLaunch(pctx, sb)
		require.Error(t, err)
		require.ErrorContains(t, err, "host mount directory not set in context")
	})
}
