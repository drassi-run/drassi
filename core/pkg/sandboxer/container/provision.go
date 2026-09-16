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

	"drassi.run/core/config"
	"drassi.run/core/pkg/container/types"
	. "drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/store/oci"
)

func NewProvisioner(store ocistore.Manager, runtimes map[string]*config.Runtime) *Provisioner[*types.ContainerSpec] {
	if len(runtimes) == 0 {
		return nil
	}
	return New[*types.ContainerSpec](
		runtimes,
		Pull[*types.ContainerSpec](store),
		Mount[*types.ContainerSpec](store),
		AddBindMount(),
	)
}

// AddBindMount returns an Operation that appends a runtime bind mount to types.ContainerSpec.Mounts.
func AddBindMount() Operation[*types.ContainerSpec] {
	return addBindMountOp{}
}

type addBindMountOp struct {
	Noop[*types.ContainerSpec]
}

func (op addBindMountOp) Name() string { return "container/bind-mount" }

func (op addBindMountOp) PreLaunch(pctx *Context, spec *types.ContainerSpec) (*types.ContainerSpec, error) {
	if spec == nil {
		return nil, errors.New("container spec cannot be nil")
	}
	hostMountDir, ok := pctx.Get(KeyHostMountDir)
	if !ok {
		return spec, fmt.Errorf("host mount directory not set in context")
	}

	sourcePath := hostMountDir
	if pctx.Config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, pctx.Config.Subpath)
	}

	spec.Mounts = append(spec.Mounts, &types.Mount{
		Type:     "bind",
		Source:   sourcePath,
		Target:   pctx.TargetDir,
		ReadOnly: pctx.Config.ReadOnly,
	})
	return spec, nil
}
