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
	"sync"

	"drassi.run/core/config"
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	ocistore "drassi.run/core/pkg/store/oci"
	xfs "drassi.run/core/util/fs"
	xpath "drassi.run/core/util/path"
	xstring "drassi.run/core/util/string"
)

func init() {
	sandboxer.Register(config.ProviderHost, DefaultConfig, NewFactory)
}

func DefaultConfig() *Config {
	return &Config{
		RootDir:    "/tmp",
		RuntimeDir: "/opt/drassi",
	}
}

func NewFactory(cfg *Config) sandboxer.Factory {
	f := &factory{cfg: cfg}
	f.create = sync.OnceValues(f.doCreate)
	return f
}

type factory struct {
	create func() (sandboxer.Engine, error)

	cfg      *Config
	store    ocistore.Manager
	runtimes map[string]*config.Runtime
}

func (f *factory) ProvisionRuntime(store ocistore.Manager, config map[string]*config.Runtime) {
	f.store = store
	f.runtimes = config
}

func (f *factory) Create() (sandboxer.Engine, error) {
	return f.create()
}

func (f *factory) doCreate() (sandboxer.Engine, error) {
	var p *provision.Provisioner[*sandboxer.LaunchRequest]
	if len(f.runtimes) > 0 && f.store != nil {
		targetDirFn := func(name string) string {
			return filepath.Join(f.cfg.RuntimeDir, name)
		}
		p = provision.New[*sandboxer.LaunchRequest](
			f.runtimes,
			targetDirFn,
			provision.Pull[*sandboxer.LaunchRequest](f.store),
			provision.Mount[*sandboxer.LaunchRequest](f.store, ocistore.WithWritable(true)),
			Symlink[*sandboxer.LaunchRequest](),
		)
	}
	return New(f.cfg, p)
}

type Config struct {
	RootDir    string `toml:"root_dir" json:"rootDir"`
	RuntimeDir string `toml:"runtime_dir,omitempty" json:"runtimeDir,omitempty"`
}

type engine struct {
	Config
	provisioner *provision.Provisioner[*sandboxer.LaunchRequest]
}

func New(config *Config, p ...*provision.Provisioner[*sandboxer.LaunchRequest]) (sandboxer.Engine, error) {
	if d, err := xpath.ResolveDir(config.RootDir); err != nil {
		return nil, err
	} else {
		config.RootDir = d
	}

	if err := os.MkdirAll(config.RootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	var prov *provision.Provisioner[*sandboxer.LaunchRequest]
	if len(p) > 0 && p[0] != nil {
		prov = p[0]
	} else {
		prov = provision.New[*sandboxer.LaunchRequest](nil, nil)
	}

	return &engine{
		Config:      *config,
		provisioner: prov,
	}, nil
}

func (e *engine) Close() error {
	return nil
}

func (e *engine) launch(ctx context.Context, req *sandboxer.LaunchRequest) (sandboxer.Sandbox, error) {
	sandboxDir := e.sandboxDir(req)
	sandboxDir = filepath.Join(e.RootDir, sandboxDir)

	sb, err := newSandbox(sandboxDir)
	if err != nil {
		return nil, err
	}
	sb.layout.Runtimes = e.RuntimeDir
	if sb.layout.Runtimes != "" {
		if err := os.MkdirAll(sb.layout.Runtimes, xfs.DirPerm); err != nil {
			_ = sb.Terminate(ctx)
			return nil, err
		}
	}
	return sb, nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	sb, err := e.provisioner.Launch(ctx, e.launch)(ctx, req)
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
