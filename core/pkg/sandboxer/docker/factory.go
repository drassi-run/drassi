/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/store/oci"
)

func init() {
	sandboxer.Register(config.ProviderDocker, DefaultConfig, NewFactory)
}

type Config struct {
	Endpoint string              `toml:"endpoint" json:"endpoint,omitempty"`
	Template *container.Template `toml:"template" json:"template,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint: "unix:///var/run/docker.sock",
		Template: &container.Template{
			Image: container.DefaultImage,
		},
	}
}

type factory struct {
	create   func() (sandboxer.Engine, error)
	cfg      *Config
	store    ocistore.Manager
	runtimes map[string]*config.Runtime
}

func NewFactory(cfg *Config) sandboxer.Factory {
	f := &factory{cfg: cfg}
	f.create = sync.OnceValues(f.doCreate)
	return f
}

func (f *factory) SetOciStore(store ocistore.Manager) {
	f.store = store
}

func (f *factory) ProvisionRuntime(cfg map[string]*config.Runtime) {
	f.runtimes = cfg
}

func (f *factory) Create() (sandboxer.Engine, error) {
	return f.create()
}

func (f *factory) doCreate() (sandboxer.Engine, error) {
	prov := container.NewProvisioner(f.store, f.runtimes)
	return New(f.cfg, prov)
}
