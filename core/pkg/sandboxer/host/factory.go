/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/store/oci"
)

func init() {
	sandboxer.Register(config.ProviderHost, DefaultConfig, NewFactory)
}

func DefaultConfig() *Config {
	return &Config{RootDir: "/tmp"}
}

type Config struct {
	RootDir string `toml:"root_dir" json:"rootDir"`
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
		prov = provision.New[string](
			f.runtimes,
			provision.Pull[string](f.store),
			provision.Mount[string](f.store),
			Symlink[string](),
		)
	}
	e, err := New(f.cfg, prov)
	if err != nil {
		return nil, err
	}
	return container.WithContainers(e), nil
}
