/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	coreconfig "drassi.run/core/config"
)

type Config = coreconfig.Config[Runner]

var DefaultConfig = coreconfig.DefaultConfig[Runner]

type Runner struct {
	Name                  string   `toml:"name" json:"name"`
	UUID                  string   `toml:"uuid" json:"uuid"`
	Token                 string   `toml:"token" json:"token"`
	Address               string   `toml:"address" json:"address"`
	InsecureSkipTLSVerify bool     `toml:"insecure_skip_tls_verify" json:"insecureSkipTLSVerify,omitempty"`
	RunnerLabels          []string `toml:"runner_labels" json:"runnerLabels,omitempty"`
	Concurrency           int      `toml:"concurrency" json:"concurrency,omitempty"`
}
