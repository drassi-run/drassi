/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import "github.com/pelletier/go-toml/v2/unstable"

type Config[R any] struct {
	Runner       *R                    `toml:"runner" json:"runner"`
	Sandboxers   map[string]*Sandboxer `toml:"sandboxers" json:"sandboxers"`
	UseSandboxer string                `toml:"use_sandboxer" json:"use_sandboxer"`
	Runtimes     map[string]*Runtime   `toml:"runtimes" json:"runtimes"`
}

type Sandboxer struct {
	Provider string              `toml:"provider" json:"provider"`
	Config   unstable.RawMessage `toml:",inline" json:",embed"`
}

type Runtime struct {
	Alias      []string `toml:"alias" json:"alias,omitempty"` // (optional) runtime alias (e.g. node20, node22, python3.13,...)
	Image      string   `toml:"image" json:"image,omitempty"`
	Subpath    string   `toml:"subpath" json:"subpath,omitempty"`       // (optional) subpath in image
	Executable string   `toml:"executable" json:"executable,omitempty"` // relative path to the binary
	Cmd        []string `toml:"cmd" json:"cmd,omitempty"`
	Paths      []string `toml:"paths" json:"paths,omitempty"` // (optional) extra directories added to PATH variable, MUST be relative
	ReadOnly   bool     `toml:"read_only" json:"read_only,omitempty"`
}

func DefaultConfig[R any]() *Config[R] {
	return &Config[R]{
		Runner: nil,
		Sandboxers: map[string]*Sandboxer{
			"host":   {Provider: ProviderHost},
			"docker": {Provider: ProviderDocker},
		},
		UseSandboxer: "host",
	}
}

const (
	ProviderDocker = "docker"
	ProviderHost   = "host"
	ProviderIncus  = "incus"
)
