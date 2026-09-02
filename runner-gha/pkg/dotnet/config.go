/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package dotnet

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	ghaconfig "drassi.run/gha-runner/config"
)

const (
	RunnerConfigFile               = ".runner"
	CredentialsConfigFile          = ".credentials"
	CredentialsRsaParamsConfigFile = ".credentials_rsaparams"
)

type Configuration struct {
	Runner               *Runner
	Credentials          *Credentials
	CredentialsRsaParams *PrivateKey
}

// Runner is model for ".runner" file
// - https://github.com/actions/runner/blob/main/src/Runner.Common/ConfigurationStore.cs#L15
// - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/ConfigurationStore.cs#L289
// - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/HostContext.cs#L232-L238
type Runner struct {
	AgentId              int64  `json:"agentId,omitempty"`
	AgentName            string `json:"agentName,omitempty"`
	PoolId               int64  `json:"poolId,omitempty"`
	PoolName             string `json:"poolName,omitempty"`
	Ephemeral            bool   `json:"ephemeral,omitempty"`
	ServerUrl            string `json:"serverUrl,omitempty"`
	ServerUrlV2          string `json:"serverUrlV2,omitempty"`
	GitHubUrl            string `json:"gitHubUrl,omitempty"`
	WorkFolder           string `json:"workFolder,omitempty"`
	DisableUpdate        bool   `json:"disableUpdate,omitempty"`
	UseV2Flow            bool   `json:"useV2Flow,omitempty"`
	SkipSessionRecover   bool   `json:"skipSessionRecover,omitempty"`
	MonitorSocketAddress string `json:"monitorSocketAddress,omitempty"`
}

// Credentials are model for ".credentials" file
// - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/CredentialData.cs#L6
// - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Common/HostContext.cs#L224-L230
// - https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/Configuration/ConfigurationManager.cs#L357-L375
type Credentials struct {
	Scheme string          `json:"scheme"`
	Data   CredentialsData `json:"data"`
}

type CredentialsData struct {
	ClientId         string `json:"clientId"`
	AuthorizationUrl string `json:"authorizationUrl"`
	RequireFips      string `json:"requireFipsCryptography,omitempty"`
}

type PublicKey struct {
	Exponent []byte `json:"exponent,omitempty"`
	Modulus  []byte `json:"modulus,omitempty"`
}

// PrivateKey RSA params is stored in ".credentials_rsaparams" file
// https://github.com/actions/runner/blob/v2.323.0/src/Runner.Listener/Configuration/RSAFileKeyManager.cs#L16
type PrivateKey struct {
	D        []byte `json:"d,omitempty"`
	P        []byte `json:"p,omitempty"`
	Q        []byte `json:"q,omitempty"`
	DP       []byte `json:"dp,omitempty"`
	DQ       []byte `json:"dq,omitempty"`
	InverseQ []byte `json:"inverseQ,omitempty"`

	PublicKey `json:",inline"`
}

func trimBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
}

// LoadConfiguration loads .runner, .credentials, and .credentials_rsaparams from the specified directory.
func LoadConfiguration(dir string) (*Configuration, error) {
	var runner Runner
	if err := load(filepath.Join(dir, RunnerConfigFile), &runner); err != nil {
		return nil, err
	}

	var cred Credentials
	if err := load(filepath.Join(dir, CredentialsConfigFile), &cred); err != nil {
		return nil, err
	}

	var rsaParams PrivateKey
	if err := load(filepath.Join(dir, CredentialsRsaParamsConfigFile), &rsaParams); err != nil {
		return nil, err
	}

	c := &Configuration{
		Runner:               &runner,
		Credentials:          &cred,
		CredentialsRsaParams: &rsaParams,
	}
	return c, nil
}

func load[O any](file string, o *O) error {
	if b, err := os.ReadFile(file); err != nil {
		return fmt.Errorf("read %s: %w", file, err)
	} else if err = json.Unmarshal(trimBOM(b), o); err != nil {
		return fmt.Errorf("unmarshal %s: %w", file, err)
	}
	return nil
}

// EncodePrivateKey converts RSA params into a base64-encoded PKCS#1 PEM string.
func (c *Configuration) EncodePrivateKey() (string, error) {
	if c.CredentialsRsaParams == nil {
		return "", fmt.Errorf("rsa private key params not found")
	}
	rsaKey, err := c.CredentialsRsaParams.ToRsaPrivateKey()
	if err != nil {
		return "", fmt.Errorf("convert rsa key: %w", err)
	}
	pri := x509.MarshalPKCS1PrivateKey(rsaKey)
	priKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pri,
	})
	return base64.StdEncoding.EncodeToString(priKey), nil
}

// ToConfig converts .NET runner configuration to gha-runner Config.
func (c *Configuration) ToConfig(sandboxerName string) (*ghaconfig.Config, error) {
	if c.Runner == nil {
		return nil, fmt.Errorf("missing runner configuration")
	}
	if c.Credentials == nil {
		return nil, fmt.Errorf("missing credentials configuration")
	}

	encodedKey, err := c.EncodePrivateKey()
	if err != nil {
		return nil, err
	}

	runner := &ghaconfig.Runner{
		RunnerId:        int(c.Runner.AgentId),
		GroupId:         int(c.Runner.PoolId),
		RunnerName:      c.Runner.AgentName,
		GroupName:       c.Runner.PoolName,
		ServerUrl:       c.Runner.ServerUrl,
		RegistrationUrl: c.Runner.GitHubUrl,
		Authorization: ghaconfig.RunnerAuthorization{
			Url:        c.Credentials.Data.AuthorizationUrl,
			ClientId:   c.Credentials.Data.ClientId,
			PrivateKey: encodedKey,
		},
	}

	if sandboxerName == "" {
		sandboxerName = "host"
	}

	cfg := ghaconfig.DefaultConfig()
	cfg.Runner = runner
	cfg.UseSandboxer = sandboxerName

	return cfg, nil
}
