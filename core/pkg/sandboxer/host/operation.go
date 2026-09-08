/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"fmt"
	"os"
	"path/filepath"

	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
)

type symlinkOp struct {
	provision.Noop
}

// Symlink returns a provision.Operation that symlinks the runtime directory in the sandbox
// to the host mount directory.
func Symlink() provision.Operation {
	return symlinkOp{}
}

func (op symlinkOp) Name() string { return "host/symlink" }

func (op symlinkOp) PostLaunch(pctx *provision.Context, sb sandboxer.Sandbox) error {
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return fmt.Errorf("host mount directory not set in context")
	}

	runtimeDir := sb.Layout().Runtimes
	target := filepath.Join(runtimeDir, pctx.RuntimeName)
	if pctx.Config.Subpath != "" {
		hostMountDir = filepath.Join(hostMountDir, pctx.Config.Subpath)
	}

	_ = os.Remove(target) // Remove existing symlink if present
	return os.Symlink(hostMountDir, target)
}
