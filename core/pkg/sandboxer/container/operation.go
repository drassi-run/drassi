/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"fmt"
	"path/filepath"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
)

type addBindMountOp struct {
	provision.Noop
	spec *types.ContainerSpec
}

func AddBindMount(spec *types.ContainerSpec) provision.Operation {
	return &addBindMountOp{spec: spec}
}

func (op *addBindMountOp) Name() string { return "container/bind-mount" }

func (op *addBindMountOp) PreLaunch(pctx *provision.Context) error {
	hostMountDir, ok := pctx.Get(provision.KeyHostMountDir)
	if !ok {
		return fmt.Errorf("host mount directory not set in context")
	}

	sourcePath := hostMountDir
	if pctx.Config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, pctx.Config.Subpath)
	}

	op.spec.Mounts = append(op.spec.Mounts, &types.Mount{
		Type:   "bind",
		Source: sourcePath,
		Target: pctx.TargetDir,
	})
	return nil
}

type addImageMountOp struct {
	provision.Noop
	spec *types.ContainerSpec
}

func AddImageMount(spec *types.ContainerSpec) provision.Operation {
	return &addImageMountOp{spec: spec}
}

func (op *addImageMountOp) Name() string { return "container/image-mount" }

func (op *addImageMountOp) PreLaunch(pctx *provision.Context) error {
	mount := &types.Mount{
		Type:   "image",
		Source: pctx.Config.Image,
		Target: pctx.TargetDir,
	}
	if pctx.Config.Subpath != "" {
		mount.ImageOptions = &types.ImageOptions{
			Subpath: pctx.Config.Subpath,
		}
	}
	op.spec.Mounts = append(op.spec.Mounts, mount)
	return nil
}
