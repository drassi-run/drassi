/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
)

type symlinkOp[Req any] struct {
	provision.Noop[Req]
}

// Symlink returns a provision.Operation that symlinks the runtime directory in the sandbox
// to the host mount directory.
func Symlink[Req any]() provision.Operation[Req] {
	return symlinkOp[Req]{}
}

func (op symlinkOp[Req]) Name() string { return "host/symlink" }

func (op symlinkOp[Req]) PostLaunch(pctx *provision.Context, sb sandboxer.Sandbox) (sandboxer.Sandbox, error) {
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return sb, fmt.Errorf("host mount directory not set in context")
	}

	runtimeDir := sb.Layout().Runtimes
	target := filepath.Join(runtimeDir, pctx.RuntimeName)
	if pctx.Config.Subpath != "" {
		hostMountDir = filepath.Join(hostMountDir, pctx.Config.Subpath)
	}

	_ = os.Remove(target) // Remove existing symlink if present
	if err := os.Symlink(hostMountDir, target); err != nil {
		return sb, err
	}
	sb = sandboxer.AddBeforeCleanup(sb, func(context.Context) error {
		return os.Remove(target)
	})
	return sb, nil
}
