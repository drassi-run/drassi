/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"context"
	"errors"
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

func (f *factory) SetOciStore(store ocistore.Manager) {
	f.store = store
}

func (f *factory) ProvisionRuntime(config map[string]*config.Runtime) {
	f.runtimes = config
}

func (f *factory) Create() (sandboxer.Engine, error) {
	return f.create()
}

func (f *factory) doCreate() (sandboxer.Engine, error) {
	var prov *provision.Provisioner[string]
	if len(f.runtimes) > 0 {
		if f.store == nil {
			return nil, errors.New("oci store is required when runtimes are configured")
		}
		prov = provision.New[string](
			f.runtimes,
			provision.Pull[string](f.store),
			provision.Mount[string](f.store),
			Symlink[string](),
		)
	}
	return New(f.cfg, prov)
}

type Config struct {
	RootDir    string `toml:"root_dir" json:"rootDir"`
	RuntimeDir string `toml:"runtime_dir,omitempty" json:"runtimeDir,omitempty"`
}

type engine struct {
	Config
	provisioner *provision.Provisioner[string]
}

func New(config *Config, prov *provision.Provisioner[string]) (sandboxer.Engine, error) {
	if d, err := xpath.ResolveDir(config.RootDir); err != nil {
		return nil, err
	} else {
		config.RootDir = d
	}

	if err := os.MkdirAll(config.RootDir, xfs.DirPerm); err != nil {
		return nil, err
	}

	return &engine{
		Config:      *config,
		provisioner: prov,
	}, nil
}

func (e *engine) Launch(ctx context.Context, req *sandboxer.LaunchRequest) (*sandboxer.LaunchResponse, error) {
	sandboxDir := e.sandboxDir(req)
	sandboxDir = filepath.Join(e.RootDir, sandboxDir)

	launcher := e.launch
	if prov := e.provisioner; prov != nil {
		runtimeDir := filepath.Join(sandboxDir, "runtime")
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

func (e *engine) launch(ctx context.Context, sandboxDir string) (sandboxer.Sandbox, error) {
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

func (e *engine) Close() error {
	return nil
}
