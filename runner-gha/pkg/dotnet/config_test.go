/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package dotnet_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"drassi.run/gha-runner/pkg/dotnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockLayout(t *testing.T, withBOM bool) string {
	t.Helper()
	tempDir := t.TempDir()

	runnerJSON := `{
  "agentId": 100,
  "agentName": "test-runner",
  "poolId": 10,
  "poolName": "test-pool",
  "serverUrl": "https://pipelines.actions.githubusercontent.com/test-tenant/",
  "gitHubUrl": "https://github.com/example/repo",
  "workFolder": "_work",
  "useV2Flow": true,
  "serverUrlV2": "https://broker.actions.githubusercontent.com/"
}`

	credentialsJSON := `{
  "scheme": "OAuth",
  "data": {
    "clientId": "11111111-2222-3333-4444-555555555555",
    "authorizationUrl": "https://token.actions.githubusercontent.com/_apis/oauth2/token/11111111-2222-3333-4444-555555555555",
    "requireFipsCryptography": "True"
  }
}`

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	dotNetKey := dotnet.NewPrivateKey(privKey)
	rsaParamsBytes, err := json.Marshal(dotNetKey)
	require.NoError(t, err)

	runnerBytes := []byte(runnerJSON)
	if withBOM {
		runnerBytes = append([]byte("\xef\xbb\xbf"), runnerBytes...)
	}

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".runner"), runnerBytes, 0600))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".credentials"), []byte(credentialsJSON), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".credentials_rsaparams"), rsaParamsBytes, 0600))

	return tempDir
}

func TestLoadConfiguration(t *testing.T) {
	tempDir := createMockLayout(t, false)

	cfg, err := dotnet.LoadConfiguration(tempDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, int64(100), cfg.Runner.AgentId)
	assert.Equal(t, "test-runner", cfg.Runner.AgentName)
	assert.Equal(t, int64(10), cfg.Runner.PoolId)
	assert.Equal(t, "test-pool", cfg.Runner.PoolName)
	assert.Equal(t, "https://github.com/example/repo", cfg.Runner.GitHubUrl)
	assert.Equal(t, "https://pipelines.actions.githubusercontent.com/test-tenant/", cfg.Runner.ServerUrl)

	assert.Equal(t, "OAuth", cfg.Credentials.Scheme)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", cfg.Credentials.Data.ClientId)
	assert.Equal(t, "https://token.actions.githubusercontent.com/_apis/oauth2/token/11111111-2222-3333-4444-555555555555", cfg.Credentials.Data.AuthorizationUrl)

	rsaKey, err := cfg.CredentialsRsaParams.ToRsaPrivateKey()
	require.NoError(t, err)
	require.NoError(t, rsaKey.Validate())

	ghaCfg, err := cfg.ToConfig("host")
	require.NoError(t, err)
	require.NotNil(t, ghaCfg)

	assert.Equal(t, "host", ghaCfg.UseSandboxer)
	require.NotNil(t, ghaCfg.Runner)
	assert.Equal(t, 100, ghaCfg.Runner.RunnerId)
	assert.Equal(t, 10, ghaCfg.Runner.GroupId)
	assert.Equal(t, "test-runner", ghaCfg.Runner.RunnerName)
	assert.Equal(t, "test-pool", ghaCfg.Runner.GroupName)
	assert.Equal(t, "https://pipelines.actions.githubusercontent.com/test-tenant/", ghaCfg.Runner.ServerUrl)
	assert.Equal(t, "https://github.com/example/repo", ghaCfg.Runner.RegistrationUrl)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", ghaCfg.Runner.Authorization.ClientId)
	assert.Equal(t, "https://token.actions.githubusercontent.com/_apis/oauth2/token/11111111-2222-3333-4444-555555555555", ghaCfg.Runner.Authorization.Url)
	assert.NotEmpty(t, ghaCfg.Runner.Authorization.PrivateKey)
}

func TestLoadConfiguration_WithBOM(t *testing.T) {
	tempDir := createMockLayout(t, true)

	cfg, err := dotnet.LoadConfiguration(tempDir)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(100), cfg.Runner.AgentId)
}

func TestLoadConfiguration_MissingFile(t *testing.T) {
	tempDir := t.TempDir()

	_, err := dotnet.LoadConfiguration(tempDir)
	assert.Error(t, err)
}
