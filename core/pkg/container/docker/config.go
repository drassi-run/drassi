/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"drassi.run/core/pkg/container/cli"
	"drassi.run/core/pkg/container/types"
	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// [github.com/docker/docker/api/types/backend.ContainerCreateConfig]
type containerConfig struct {
	Name             string
	Config           *dockercontainer.Config
	HostConfig       *dockercontainer.HostConfig
	NetworkingConfig *dockernetwork.NetworkingConfig
	Platform         *ocispec.Platform
}

// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L602-L703
func (cc *containerConfig) From(spec *types.ContainerSpec, stdio *types.Stdio) error {
	cc.Name = spec.Name

	cc.Config = &dockercontainer.Config{
		Image:      spec.Image,
		Entrypoint: spec.Entrypoint,
		Cmd:        spec.Command,
		WorkingDir: spec.WorkingDir,
		Env:        cli.ConvertMapToKVString(spec.Environment),
		Labels:     spec.Labels,
	}

	cc.HostConfig = &dockercontainer.HostConfig{
		Annotations: spec.Annotations,
	}

	if err := cc.setResources(&spec.ContainerResource); err != nil {
		return err
	}
	cc.setStorage(&spec.ContainerStorage)
	cc.setNetwork(&spec.ContainerNetwork)
	cc.setRuntime(&spec.ContainerRuntime)
	cc.setSecurity(&spec.ContainerSecurity)
	cc.setStdio(stdio)

	return nil
}

func (cc *containerConfig) setStdio(stdio *types.Stdio) {
	c := cc.Config

	c.Tty = stdio.Tty
	c.OpenStdin = stdio.Interactive
	c.AttachStdin = stdio.AttachStdin()
	c.AttachStdout = stdio.AttachStdout()
	c.AttachStderr = stdio.AttachStderr()

	// When allocating stdin in attached mode, close stdin at client disconnect
	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L715-L718
	c.StdinOnce = stdio.Interactive && stdio.AttachStdin()
}

type containerSpec struct {
	Spec *types.ContainerSpec
}

func (cs *containerSpec) From(info dockertypes.ContainerJSON) error {
	c, hc := info.Config, info.HostConfig

	cs.Spec = &types.ContainerSpec{
		Name:        info.Name,
		Image:       c.Image,
		Command:     c.Cmd,
		Entrypoint:  c.Entrypoint,
		WorkingDir:  c.WorkingDir,
		Environment: cli.ConvertKVStringsToMap(c.Env),
		Labels:      c.Labels,
		Annotations: hc.Annotations,
	}

	if err := cs.setNetwork(info); err != nil {
		return err
	}
	if err := cs.setStorage(info); err != nil {
		return err
	}
	cs.setResources(hc)
	cs.setRuntime(c, hc)
	cs.setSecurity(c, hc)

	return nil
}
