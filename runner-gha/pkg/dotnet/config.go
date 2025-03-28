/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package dotnet

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
	Ephemeral            string `json:"ephemeral,omitempty,omitempty"`
	ServerUrl            string `json:"serverUrl,omitempty"`
	ServerUrlV2          string `json:"serverUrlV2,omitempty"`
	GitHubUrl            string `json:"gitHubUrl,omitempty"`
	WorkFolder           string `json:"workFolder,omitempty"`
	DisableUpdate        string `json:"disableUpdate,omitempty"`
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
