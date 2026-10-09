/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"context"
	"errors"
	"path/filepath"

	"drassi.run/core/config"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/store/oci"
)

func NewProvisioner(store ocistore.Manager, runtimes map[string]*config.Runtime) (*provision.Provisioner[*types.ContainerSpec], error) {
	if len(runtimes) == 0 {
		return nil, nil
	}
	if store == nil {
		return nil, errors.New("oci store required for runtimes")
	}
	factory := func(runtimeDir string, name string, rt config.Runtime) provision.Pipeline[*types.ContainerSpec] {
		targetDir := filepath.Join(DefaultLayout.Runtimes(), name)
		mountDir := filepath.Join(runtimeDir, name)
		state := new(provision.State)
		return provision.Pipeline[*types.ContainerSpec]{
			provision.Pull[*types.ContainerSpec](store, state, rt.Image),
			provision.Mount[*types.ContainerSpec](store, state,
				ocistore.WithTarget(mountDir),
				ocistore.WithWritable(!rt.ReadOnly),
			),
			AddBindMount(mountDir, targetDir, rt),
		}
	}
	return provision.New(runtimes, factory), nil
}

// AddBindMount returns an Operation that appends a runtime bind mount to types.ContainerSpec.Mounts.
func AddBindMount(sourceDir, targetDir string, cfg config.Runtime) provision.Operation[*types.ContainerSpec] {
	return &addBindMountOp{
		sourceDir: sourceDir,
		targetDir: targetDir,
		config:    cfg,
	}
}

type addBindMountOp struct {
	provision.Noop[*types.ContainerSpec]
	sourceDir string
	targetDir string
	config    config.Runtime
}

func (op *addBindMountOp) Name() string { return "container/bind-mount" }

func (op *addBindMountOp) PreLaunch(_ context.Context, spec *types.ContainerSpec) (*types.ContainerSpec, error) {
	if spec == nil {
		return nil, errors.New("container spec cannot be nil")
	}

	sourcePath := op.sourceDir
	if op.config.Subpath != "" {
		sourcePath = filepath.Join(sourcePath, op.config.Subpath)
	}

	spec.Mounts = append(spec.Mounts, &types.Mount{
		Type:     "bind",
		Source:   sourcePath,
		Target:   op.targetDir,
		ReadOnly: op.config.ReadOnly,
	})
	return spec, nil
}
