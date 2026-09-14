/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"context"
	"os"
	"path/filepath"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
)

func New(config *Config, prov *provision.Provisioner[string]) (e sandboxer.Engine, err error) {
	rootDir := config.RootDir
	if rootDir, err = xpath.ResolveDir(rootDir); err != nil {
		return
	}

	if err = os.MkdirAll(config.RootDir, xfs.DirPerm); err != nil {
		return
	}

	e = &engine{
		rootDir:     rootDir,
		provisioner: prov,
	}
	return
}

type engine struct {
	rootDir     string
	provisioner *provision.Provisioner[string]
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	sandboxDir := req.Forge.StandardPath()
	sandboxDir = filepath.Join(e.rootDir, sandboxDir)

	launcher := func(_ context.Context, dir string) (sandboxer.Sandbox, error) {
		return newSandbox(dir)
	}
	if prov := e.provisioner; prov != nil {
		runtimeDir := filepath.Join(sandboxDir, "runtimes")
		launcher = prov.Launch(runtimeDir, launcher)
	}
	sb, err := launcher(ctx, sandboxDir)
	if err != nil {
		return nil, err
	}

	client, err := docker.New()
	if err != nil {
		_ = sb.Terminate(ctx)
		return nil, err
	}
	client = c.WithTelemetry(client)

	b := container.NewBootstrapper(client)
	resp, err := b.Bootstrap(ctx, sb, req)
	if err != nil {
		_ = sb.Terminate(ctx)
		return nil, err
	}
	return resp, nil
}

func (e *engine) Close() error {
	return nil
}
