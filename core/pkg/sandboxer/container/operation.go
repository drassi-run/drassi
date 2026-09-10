/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"errors"
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
)

type addBindMountOp struct {
	provision.Noop[*types.ContainerSpec]
}

// AddBindMount returns an Operation that appends a runtime bind mount to types.ContainerSpec.Mounts.
func AddBindMount() provision.Operation[*types.ContainerSpec] {
	return &addBindMountOp{}
}

func (op *addBindMountOp) Name() string { return "container/bind-mount" }

func (op *addBindMountOp) PreLaunch(pctx *provision.Context, spec *types.ContainerSpec) (*types.ContainerSpec, error) {
	if spec == nil {
		return nil, errors.New("container spec cannot be nil")
	}
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return spec, fmt.Errorf("host mount directory not set in context")
	}

	sourcePath := hostMountDir
	if pctx.Config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, pctx.Config.Subpath)
	}

	spec.Mounts = append(spec.Mounts, &types.Mount{
		Type:   "bind",
		Source: sourcePath,
		Target: pctx.TargetDir,
	})
	return spec, nil
}
