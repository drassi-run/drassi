/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"fmt"
	"maps"
	"strings"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/container/types"
)

type Option func(*types.ContainerSpec) error

func SetLabels(labels map[string]string) Option {
	return func(spec *types.ContainerSpec) error {
		// set labels for container
		if spec.Labels == nil {
			spec.Labels = maps.Clone(labels)
		} else {
			maps.Copy(spec.Labels, labels)
		}

		// set labels for volumes
		for _, vol := range spec.Mounts {
			if vol.Type != "volume" {
				continue
			}
			if vol.VolumeOptions == nil {
				vol.VolumeOptions = new(types.VolumeOptions)
			}
			if opts := vol.VolumeOptions; opts.Labels == nil {
				opts.Labels = maps.Clone(labels)
			} else {
				maps.Copy(opts.Labels, labels)
			}
		}

		return nil
	}
}

func SetCmd(entrypoint, command []string) Option {
	return func(spec *types.ContainerSpec) error {
		if len(entrypoint) > 0 {
			spec.Entrypoint = entrypoint
			spec.Command = nil
		}
		if len(command) > 0 {
			spec.Command = command
		}
		return nil
	}
}

func SetNetwork(netId string) Option {
	return func(spec *types.ContainerSpec) error {
		switch len(spec.Endpoints) {
		case 0:
			endpoint := &types.Endpoint{Target: netId}
			spec.Endpoints = append(spec.Endpoints, endpoint)
		case 1:
			if endpoint := spec.Endpoints[0]; endpoint.Target != "" {
				return fmt.Errorf("can't overwrite non-default network %q", endpoint.Target)
			} else {
				endpoint.Target = netId
			}
		default:
			return fmt.Errorf("only one network per container")
		}
		return nil
	}
}

func AddMount(mounts ...*types.Mount) Option {
	return func(spec *types.ContainerSpec) error {
		spec.Mounts = append(spec.Mounts, mounts...)
		return nil
	}
}

func MountVolume(volume, target string) Option {
	if target == "" {
		target = DefaultJobDir
	}
	return func(spec *types.ContainerSpec) error {
		spec.Mounts = append(spec.Mounts, &types.Mount{
			Type:   "volume",
			Source: volume,
			Target: target,
		})
		return nil
	}
}

func MountApiSocket(c container.Engine) Option {
	socket := c.Address()
	if proto, loc, ok := strings.Cut(socket, "://"); ok {
		if proto == "unix" {
			socket = loc
		} else {
			return func(container *types.ContainerSpec) error { return nil }
		}
	}
	return func(spec *types.ContainerSpec) error {
		m := &types.Mount{
			Type:   "bind",
			Source: socket,
			Target: socket,
		}
		spec.Mounts = append(spec.Mounts, m)
		return nil
	}
}

func SetWorkdir(dir string) Option {
	return func(spec *types.ContainerSpec) error {
		if spec.WorkingDir != "" {
			spec.WorkingDir = dir
		}
		return nil
	}
}

func SetCIEnv() Option {
	return func(spec *types.ContainerSpec) error {
		if spec.Environment == nil {
			spec.Environment = make(map[string]string)
		}
		spec.Environment["CI"] = "true"
		spec.Environment["GITHUB_ACTIONS"] = "true"

		return nil
	}
}
