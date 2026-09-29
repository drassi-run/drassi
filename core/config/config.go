/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import "github.com/pelletier/go-toml/v2/unstable"

type Config[R any] struct {
	Runner    *R                  `toml:"runner" json:"runner"`
	Sandboxer *Sandboxer          `toml:"sandboxer" json:"sandboxer"`
	Runtimes  map[string]*Runtime `toml:"runtimes" json:"runtimes"`
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
		Sandboxer: &Sandboxer{
			Provider: "host",
		},
		Runtimes: map[string]*Runtime{
			"node": {
				Alias:      []string{"node20", "node22", "node24"},
				Image:      "ghcr.io/drassi-run/runtimes/node:24",
				Executable: "./bin/node",
				Cmd:        []string{"{0}"},
				Paths:      []string{"bin"},
			},
		},
	}
}
