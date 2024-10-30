package types

import "time"

type ContainerRuntime struct {
	Runtime       string
	Platform      string
	Isolation     string
	RestartPolicy *RestartPolicy
	AutoRemove    bool
	StopSignal    string
	StopTimeout   *time.Duration
	Logging       *LoggingConfig
	HealthCheck   *HealthCheckConfig
}

// RestartPolicy represents the restart policies of the container.
//   - [github.com/docker/docker/api/types/container.RestartPolicy]
//   - [github.com/docker/docker/api/types/swarm.RestartPolicy]
//   - [github.com/docker/cli/cli/compose/types.RestartPolicy]
//   - [github.com/compose-spec/compose-go/v2/types.RestartPolicy]
type RestartPolicy struct {
	Name     string // "no", "on-failure", "always", "unless-stopped"
	MaxRetry int
}

// LoggingConfig is identical with compose LoggingConfig
//   - [github.com/docker/docker/api/types/container.LogConfig]
//   - [github.com/compose-spec/compose-go/v2/types.LoggingConfig]
type LoggingConfig struct {
	Driver  string
	Options map[string]string
}

// HealthCheckConfig is identical with docker's HealthConfig
//   - [github.com/docker/docker/api/types/container.HealthConfig]
//   - [github.com/compose-spec/compose-go/v2/types.HealthCheckConfig]
//   - [github.com/containers/image/v5/manifest.Schema2HealthConfig]
type HealthCheckConfig struct {
	Test          []string
	Timeout       time.Duration
	Interval      time.Duration
	Retries       int
	StartPeriod   time.Duration
	StartInterval time.Duration
}
