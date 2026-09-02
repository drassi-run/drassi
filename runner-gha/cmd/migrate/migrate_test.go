/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package migrate_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	coreconfig "drassi.run/core/config"
	"drassi.run/gha-runner/cmd/migrate"
	ghaconfig "drassi.run/gha-runner/config"
	"drassi.run/gha-runner/pkg/dotnet"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

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

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".runner"), []byte(runnerJSON), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".credentials"), []byte(credentialsJSON), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".credentials_rsaparams"), rsaParamsBytes, 0600))
	return dir
}

func TestMigrateCommand(t *testing.T) {
	sourceDir := setupTestDir(t)
	outputFile := filepath.Join(t.TempDir(), "config.toml")

	cmd := migrate.New()
	cmd.SetArgs([]string{sourceDir, "-o", outputFile, "--sandboxer", "host"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	cfg := ghaconfig.DefaultConfig()
	err = toml.Unmarshal(content, cfg)
	require.NoError(t, err)

	assert.Equal(t, "host", cfg.UseSandboxer)
	require.NotNil(t, cfg.Runner)
	assert.Equal(t, 100, cfg.Runner.RunnerId)
	assert.Equal(t, 10, cfg.Runner.GroupId)
	assert.Equal(t, "test-runner", cfg.Runner.RunnerName)
	assert.Equal(t, "test-pool", cfg.Runner.GroupName)
	assert.Equal(t, "https://pipelines.actions.githubusercontent.com/test-tenant/", cfg.Runner.ServerUrl)
	assert.Equal(t, "https://github.com/example/repo", cfg.Runner.RegistrationUrl)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", cfg.Runner.Authorization.ClientId)
	assert.Equal(t, "https://token.actions.githubusercontent.com/_apis/oauth2/token/11111111-2222-3333-4444-555555555555", cfg.Runner.Authorization.Url)
	assert.NotEmpty(t, cfg.Runner.Authorization.PrivateKey)

	// Check sandboxer engine instantiation
	sb, ok := cfg.Sandboxers[cfg.UseSandboxer]
	require.True(t, ok)
	engine, err := coreconfig.NewSandboxerEngine(sb)
	require.NoError(t, err)
	require.NotNil(t, engine)
}

func TestMigrateCommand_DefaultArgs(t *testing.T) {
	sourceDir := setupTestDir(t)

	// Switch working dir to sourceDir for test
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(sourceDir))

	cmd := migrate.New()
	cmd.SetArgs([]string{}) // No args, should default to . and config.toml

	err = cmd.Execute()
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(sourceDir, "config.toml"))
}

func TestMigrateCommand_InvalidDir(t *testing.T) {
	emptyDir := t.TempDir()

	cmd := migrate.New()
	cmd.SetArgs([]string{emptyDir})

	err := cmd.Execute()
	assert.Error(t, err)
}
