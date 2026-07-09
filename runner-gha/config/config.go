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
	RunnerId   int      `toml:"runner_id" json:"runnerId"`
	GroupId    int      `toml:"group_id" json:"groupId"`
	RunnerName string   `toml:"runner_name" json:"runnerName,omitempty"`
	GroupName  string   `toml:"group_name" json:"groupName,omitempty"`
	Labels     []string `toml:"labels" json:"labels,omitempty"`

	ServerUrl       string              `toml:"server_url" json:"serverUrl"`
	RegistrationUrl string              `toml:"registration_url" json:"registrationUrl"`
	Authorization   RunnerAuthorization `toml:"authorization" json:"authorization,omitempty"`
}

type RunnerAuthorization struct {
	Url            string `toml:"url" json:"url"`
	ClientId       string `toml:"client_id" json:"clientId"`
	PrivateKey     string `toml:"private_key" json:"privateKey"`          // base64-encoded PEM-format private.key, mutually exclusive with PrivateKeyFile
	PrivateKeyFile string `toml:"private_key_file" json:"privateKeyFile"` // path to private.key in PEM-format, mutually exclusive with PrivateKey
	PublicKey      string `toml:"public_key" json:"publicKey"`            // base64-encoded PEM-format public.key, mutually exclusive with PublicKeyFile
	PublicKeyFile  string `toml:"public_key_file" json:"publicKeyFile"`   // path to public.key in PEM format, mutually exclusive with PublicKey
}
