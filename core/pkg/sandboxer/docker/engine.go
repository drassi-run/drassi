/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"sync"

	"drassi.run/core/config"
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	ocistore "drassi.run/core/pkg/store/oci"
	dockerclient "github.com/moby/moby/client"
)

func init() {
	sandboxer.Register(config.ProviderDocker, DefaultConfig, NewFactory)
}

type Config struct {
	Endpoint string `toml:"endpoint" json:"endpoint,omitempty"`
	Image    string `toml:"image" json:"image,omitempty"`
}

func DefaultConfig() *Config {
	return &Config{
		Image: container.DefaultImage,
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
	prov, err := container.NewProvisioner(f.store, f.runtimes)
	if err != nil {
		return nil, err
	}
	return New(f.cfg, prov)
}

func New(cfg *Config, prov *provision.Provisioner[*types.ContainerSpec]) (sandboxer.Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	opts := make([]dockerclient.Opt, 0)
	if ep := cfg.Endpoint; ep != "" {
		opts = append(opts, dockerclient.WithHost(ep))
	}

	client, err := docker.New(opts...)
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	return container.New(client, cfg.Image, prov), nil
}
