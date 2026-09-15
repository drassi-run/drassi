/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	c "drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/docker"
	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/runtime/provision"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/sandboxer/container"
	dockerclient "github.com/moby/moby/client"
)

func New(cfg *Config, prov *provision.Provisioner[*types.ContainerSpec]) (sandboxer.Engine, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	opts := make([]dockerclient.Opt, 0)
	if ep := cfg.Endpoint; ep != "" {
		opts = append(opts, dockerclient.WithHost(ep))
	}

	client, err := docker.New(opts...)
	if err != nil {
		return nil, err
	}
	client = c.WithTelemetry(client)

	return container.New(client, cfg.Template, prov), nil
}
