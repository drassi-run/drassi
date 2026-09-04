/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/util/fs"
	"drassi.run/core/util/path"
	"drassi.run/core/util/string"
)

type Config struct {
	RootDir    string `toml:"root_dir" json:"rootDir"`
	RuntimeDir string `toml:"runtime_dir,omitempty" json:"runtimeDir,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		RootDir:    "/tmp",
		RuntimeDir: "/opt/drassi",
	}
}

type engine struct {
	Config
}

func New(config *Config) (sandboxer.Engine, error) {
	if d, err := xpath.ResolveDir(config.RootDir); err != nil {
		return nil, err
	} else {
		config.RootDir = d
	}

	if err := os.MkdirAll(config.RootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	return &engine{Config: *config}, nil
}

func (e *engine) Close() error {
	return nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	sandboxDir := e.sandboxDir(req)
	sandboxDir = filepath.Join(e.RootDir, sandboxDir)

	sb, err := newSandbox(sandboxDir)
	if err != nil {
		return nil, err
	}
	sb.layout.Runtimes = e.RuntimeDir

	client, err := docker.New()
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	b := container.NewBootstrapper(client)
	return b.Bootstrap(ctx, sb, req)
}

func (e *engine) sandboxDir(req *sandboxer.LaunchRequest) string {
	var server string
	if u, err := url.Parse(req.Forge.ServerUrl); err == nil {
		server = u.Host
	}
	server = strings.ToLower(server)
	repo := strings.ToLower(req.Forge.Repository)

	workflow := strings.TrimSuffix(req.Forge.Workflow, ".yml")
	workflow = strings.TrimSuffix(workflow, ".yaml")
	workflow = xstring.Normalize(workflow)

	job := xstring.Normalize(req.Forge.Job)
	run := xstring.Normalize(req.Forge.RunId)
	attempt := xstring.Normalize(req.Forge.RunAttempt)

	path := filepath.Join(server, repo, workflow, job, run+"_"+attempt)
	return path
}
