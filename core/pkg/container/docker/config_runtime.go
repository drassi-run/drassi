/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"time"

	"drassi.run/core/pkg/container/types"
	dockercontainer "github.com/docker/docker/api/types/container"
)

func (cc *containerConfig) setRuntime(conf *types.ContainerRuntime) {
	c, hc := cc.Config, cc.HostConfig

	hc.Runtime = conf.Runtime
	hc.Isolation = dockercontainer.Isolation(conf.Isolation)
	hc.AutoRemove = conf.AutoRemove
	c.StopSignal = conf.StopSignal

	if d := conf.StopTimeout; d != nil {
		c.StopTimeout = new(int(d.Seconds()))
	}

	if restart := conf.RestartPolicy; restart != nil {
		hc.RestartPolicy = dockercontainer.RestartPolicy{
			Name:              dockercontainer.RestartPolicyMode(restart.Name),
			MaximumRetryCount: restart.MaxRetry,
		}
	}

	if log := conf.Logging; log != nil {
		hc.LogConfig = dockercontainer.LogConfig{
			Type:   log.Driver,
			Config: log.Options,
		}
	}

	if check := conf.HealthCheck; check != nil {
		c.Healthcheck = &dockercontainer.HealthConfig{
			Test:          check.Test,
			Timeout:       check.Timeout,
			Interval:      check.Interval,
			Retries:       check.Retries,
			StartPeriod:   check.StartPeriod,
			StartInterval: check.StartInterval,
		}
	}
}

func (cs *containerSpec) setRuntime(c *dockercontainer.Config, hc *dockercontainer.HostConfig) {
	cs.Spec.ContainerRuntime = types.ContainerRuntime{
		Runtime:    hc.Runtime,
		Isolation:  string(hc.Isolation),
		AutoRemove: hc.AutoRemove,
		StopSignal: c.StopSignal,
	}
	r := &cs.Spec.ContainerRuntime

	if s := c.StopTimeout; s != nil {
		r.StopTimeout = new(time.Duration(*s) * time.Second)
	}

	if restart := hc.RestartPolicy; !restart.IsNone() {
		r.RestartPolicy = &types.RestartPolicy{
			Name:     string(restart.Name),
			MaxRetry: restart.MaximumRetryCount,
		}
	}

	if log := hc.LogConfig; log.Type != "" && log.Config != nil {
		r.Logging = &types.LoggingConfig{
			Driver:  log.Type,
			Options: log.Config,
		}
	}

	if check := c.Healthcheck; check != nil {
		r.HealthCheck = &types.HealthCheckConfig{
			Test:          check.Test,
			Timeout:       check.Timeout,
			Interval:      check.Interval,
			Retries:       check.Retries,
			StartPeriod:   check.StartPeriod,
			StartInterval: check.StartInterval,
		}
	}
}
