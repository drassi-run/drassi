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
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"drassi.run/core/util/sync"
)

func New(jobsDir string, prov *provision.Provisioner[string]) (e sandboxer.Engine, err error) {
	if jobsDir, err = xpath.ResolveDir(jobsDir); err != nil {
		return
	}

	if err = os.MkdirAll(jobsDir, xfs.DirPerm); err != nil {
		return
	}

	e = &engine{
		jobsDir:     jobsDir,
		provisioner: prov,
	}
	return
}

type engine struct {
	jobsDir     string
	provisioner *provision.Provisioner[string]
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	jobDir := filepath.Join(e.jobsDir, req.Uid)

	launcher := func(_ context.Context, dir string) (sandboxer.Sandbox, error) {
		return newSandbox(dir)
	}
	if prov := e.provisioner; prov != nil {
		runtimesDir := filepath.Join(jobDir, "runtimes")
		launcher = prov.Launch(runtimesDir, launcher)
	}
	launcher = sandboxer.SetupJobDir(jobDir, launcher,
		sandboxer.WithForge(req.Forge),
		sandboxer.WithSandboxId("host", req.Uid),
	)
	sb, err := launcher(ctx, jobDir)
	if err != nil {
		return nil, err
	}

	ce := func(context.Context) (c.Engine, error) {
		client, err := docker.New()
		if err != nil {
			return nil, err
		}
		return c.WithTelemetry(client), nil
	}
	resp := &sandboxer.LaunchResponse{
		Sandbox:         sb,
		Mounter:         newMounter(jobDir),
		ContainerEngine: xsync.Singleton(ce),
	}
	return resp, nil
}

func (e *engine) Close() error {
	return nil
}
