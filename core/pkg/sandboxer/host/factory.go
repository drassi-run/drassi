/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package host

import (
	"errors"
	"path/filepath"
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
	return new(Config)
}

type Config struct{}

func NewFactory(_ *Config) sandboxer.Factory {
	f := new(factory)
	f.create = sync.OnceValues(f.doCreate)
	return f
}

type factory struct {
	create func() (sandboxer.Engine, error)

	rootDir  string
	store    ocistore.Manager
	runtimes map[string]*config.Runtime
}

func (f *factory) RootDir(d string) {
	f.rootDir = d
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
			return nil, errors.New("oci store required for runtimes")
		}
		fact := func(runtimeDir string, name string, rt config.Runtime) provision.Pipeline[string] {
			mountDir := filepath.Join(runtimeDir, name)
			state := new(provision.State)
			return provision.Pipeline[string]{
				provision.Pull[string](f.store, state, rt.Image),
				provision.Mount[string](f.store, state,
					ocistore.WithTarget(mountDir),
					ocistore.WithWritable(!rt.ReadOnly),
				),
			}
		}
		prov = provision.New(f.runtimes, fact)
	}
	e, err := New(f.rootDir, prov)
	if err != nil {
		return nil, err
	}
	return container.WithContainers(e), nil
}
