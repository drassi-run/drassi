/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package incus

import (
	"maps"
	"slices"
	"sync"

	"drassi.run/core/config"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	ocistore "drassi.run/core/pkg/store/oci"
)

func init() {
	sandboxer.Register(config.ProviderIncus, DefaultConfig, NewFactory)
}

func DefaultConfig() *Config {
	return &Config{
		Endpoint: "unix:///var/lib/incus/unix.socket",
		Template: Template{
			Image:     "ubuntu:latest",
			Ephemeral: true,
		},
	}
}

type Config struct {
	Endpoint string   `toml:"endpoint" json:"endpoint"`
	Template Template `toml:"template" json:"template"`
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
	var prov *provision.Provisioner[*Template]
	if len(f.runtimes) > 0 {
		prov = provision.New[*Template](
			f.runtimes,
			provision.Pull[*Template](f.store),
			provision.Mount[*Template](f.store),
			AddDiskDevice(),
		)
	}
	e, err := New(f.cfg, prov)
	if err != nil {
		return nil, err
	}
	return container.WithContainers(e), nil
}

// Template for create incus VM
// [github.com/lxc/incus/v6/shared/api.InstancesPost]
type Template struct {
	Name string `toml:"name,omitempty" json:"name,omitempty"`

	// OCI image name, e.g: ghcr.io/drassi-run/ubuntu:22.04
	Image string `toml:"source" json:"image"`

	// Instance architecture, e.g: x86_64
	Architecture string `toml:"architecture" json:"architecture,omitempty"`

	// Cloud instance size (AWS, GCP, Azure, ...) to emulate with limits
	// Example: t1.micro
	InstanceSize string `toml:"instance_size" json:"instance_size,omitempty"`

	// List of profiles applied to the instance
	// Example: ["default"]
	Profiles []string `toml:"profiles" json:"profiles,omitempty"`

	// Instance configuration (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"security.nesting": "true"}
	Config map[string]string `toml:"config" json:"config,omitempty"`

	// Instance devices (see https://linuxcontainers.org/incus/docs/main/instances/)
	// Example: {"root": {"type": "disk", "pool": "default", "path": "/"}}
	Devices map[string]map[string]string `toml:"devices" json:"devices,omitempty"`

	// Whether the instance is ephemeral (deleted on shutdown)
	// Example: false
	Ephemeral bool `toml:"ephemeral" json:"ephemeral,omitempty"`
}

func (t *Template) Clone() *Template {
	if t == nil {
		return nil
	}
	c := *t
	c.Profiles = slices.Clone(t.Profiles)
	c.Config = maps.Clone(t.Config)
	if t.Devices != nil {
		c.Devices = make(map[string]map[string]string, len(t.Devices))
		for k, v := range t.Devices {
			c.Devices[k] = maps.Clone(v)
		}
	}
	return &c
}
