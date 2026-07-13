/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	"drassi.run/core/pkg/sandboxer/host"
	"drassi.run/core/pkg/sandboxer/incus"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

type Config[R any] struct {
	Runner       *R                    `toml:"runner" json:"runner"`
	Sandboxers   map[string]*Sandboxer `toml:"sandboxers" json:"sandboxers"`
	UseSandboxer string                `toml:"use_sandboxer"`
}

type Sandboxer struct {
	Provider string              `toml:"provider" json:"provider"`
	Config   unstable.RawMessage `toml:",inline" json:",inline"`
}

func DefaultConfig[R any]() *Config[R] {
	return &Config[R]{
		Runner: nil,
		Sandboxers: map[string]*Sandboxer{
			"host":   {Provider: ProviderHost},
			"docker": {Provider: ProviderContainer},
		},
		UseSandboxer: "host",
	}
}

const (
	ProviderContainer = "container"
	ProviderHost      = "host"
	ProviderIncus     = "incus"
)

func NewSandboxerEngine(config *Sandboxer) (sandboxer.Engine, error) {
	provider := strings.ToLower(config.Provider)
	switch provider {
	case ProviderContainer:
		cfg := container.DefaultConfig()
		if err := toml.Unmarshal(config.Config, cfg); err != nil {
			return nil, err
		}
		return container.New(cfg)
	case ProviderHost:
		cfg := host.DefaultConfig()
		if err := toml.Unmarshal(config.Config, cfg); err != nil {
			return nil, err
		}
		return host.New(cfg)
	case ProviderIncus:
		cfg := incus.DefaultConfig()
		if err := toml.Unmarshal(config.Config, cfg); err != nil {
			return nil, err
		}
		return incus.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported sandboxer provider %q", config.Provider)
	}
}
