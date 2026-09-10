/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"fmt"

	"drassi.run/core/config"
	"drassi.run/core/pkg/sandboxer"
)

type Provider interface {
	Get(nameOrAlias string) (Runtime, error)
}

type provider struct {
	runtimes map[string]Runtime
	aliases  map[string]string
}

func NewProvider(sandbox sandboxer.Sandbox, runtimeConfigs map[string]*config.Runtime) (Provider, error) {
	p := &provider{
		runtimes: make(map[string]Runtime, len(runtimeConfigs)),
		aliases:  make(map[string]string),
	}

	for name, cfg := range runtimeConfigs {
		if rt, err := NewRuntime(name, sandbox, cfg); err != nil {
			return nil, err
		} else {
			p.runtimes[name] = rt
		}

		for _, alias := range cfg.Alias {
			if existing, ok := p.aliases[alias]; ok {
				return nil, fmt.Errorf("duplicate runtime alias %q for %q (already defined in %q)", alias, name, existing)
			}
			if _, ok := runtimeConfigs[alias]; ok {
				return nil, fmt.Errorf("runtime alias %q for %q conflicts with existing runtime name", alias, name)
			}
			p.aliases[alias] = name
		}
	}

	return p, nil
}

func (p *provider) Get(nameOrAlias string) (Runtime, error) {
	name := nameOrAlias
	if n, ok := p.aliases[name]; ok {
		name = n
	}
	if rt, ok := p.runtimes[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("unsupported runtime %q", nameOrAlias)
}
