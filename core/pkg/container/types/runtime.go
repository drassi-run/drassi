/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import "time"

type ContainerRuntime struct {
	Runtime       string             `json:"runtime,omitempty"`
	Platform      string             `json:"platform,omitempty"`
	Isolation     string             `json:"isolation,omitempty"`
	RestartPolicy *RestartPolicy     `json:"restart,omitempty"`
	AutoRemove    bool               `json:"auto_remove,omitempty"`
	StopSignal    string             `json:"stop_signal,omitempty"`
	StopTimeout   *time.Duration     `json:"stop_timeout,omitempty"`
	Logging       *LoggingConfig     `json:"logging,omitempty"`
	HealthCheck   *HealthCheckConfig `json:"healthcheck,omitempty"`
}

// RestartPolicy represents the restart policies of the container.
//   - [github.com/moby/moby/api/types/container.RestartPolicy]
//   - [github.com/moby/moby/api/types/swarm.RestartPolicy]
//   - [github.com/docker/cli/cli/compose/types.RestartPolicy]
//   - [github.com/compose-spec/compose-go/v2/types.RestartPolicy]
type RestartPolicy struct {
	Name     string `json:"name,omitempty"` // "no", "on-failure", "always", "unless-stopped"
	MaxRetry int    `json:"max_retry,omitempty"`
}

// LoggingConfig is identical with compose LoggingConfig
//   - [github.com/moby/moby/api/types/container.LogConfig]
//   - [github.com/compose-spec/compose-go/v2/types.LoggingConfig]
type LoggingConfig struct {
	Driver  string            `json:"driver,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// HealthCheckConfig is identical with docker's HealthConfig
//   - [github.com/moby/moby/api/types/container.HealthConfig]
//   - [github.com/compose-spec/compose-go/v2/types.HealthCheckConfig]
//   - [github.com/containers/image/v5/manifest.Schema2HealthConfig]
type HealthCheckConfig struct {
	Test          []string      `json:"test,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`
	Interval      time.Duration `json:"interval,omitempty"`
	Retries       int           `json:"retries,omitempty"`
	StartPeriod   time.Duration `json:"start_period,omitempty"`
	StartInterval time.Duration `json:"start_interval,omitempty"`
}
