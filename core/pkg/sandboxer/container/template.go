/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"maps"
	"slices"

	"drassi.run/core/pkg/container/types"
	"drassi.run/core/pkg/model"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

var _ unstable.Unmarshaler = (*Template)(nil)

type Template types.ContainerSpec

func (t *Template) UnmarshalTOML(data []byte) error {
	var rawMap any
	if err := toml.Unmarshal(data, &rawMap); err != nil {
		return err
	}
	return model.Decode(rawMap, (*types.ContainerSpec)(t))
}

func (t *Template) Copy() *Template {
	if t == nil {
		return nil
	}
	spec := (*types.ContainerSpec)(t)
	clone := *spec
	if spec.Command != nil {
		clone.Command = slices.Clone(spec.Command)
	}
	if spec.Entrypoint != nil {
		clone.Entrypoint = slices.Clone(spec.Entrypoint)
	}
	if spec.Environment != nil {
		clone.Environment = maps.Clone(spec.Environment)
	}
	if spec.Labels != nil {
		clone.Labels = maps.Clone(spec.Labels)
	}
	if spec.Annotations != nil {
		clone.Annotations = maps.Clone(spec.Annotations)
	}
	if spec.Devices != nil {
		clone.Devices = slices.Clone(spec.Devices)
	}
	if spec.DeviceCgroupRules != nil {
		clone.DeviceCgroupRules = slices.Clone(spec.DeviceCgroupRules)
	}
	if spec.Exposes != nil {
		clone.Exposes = make([]*types.Port, len(spec.Exposes))
		for i, p := range spec.Exposes {
			if p != nil {
				cp := *p
				clone.Exposes[i] = &cp
			}
		}
	}
	if spec.Publish != nil {
		clone.Publish = make([]*types.PortBinding, len(spec.Publish))
		for i, pb := range spec.Publish {
			if pb != nil {
				cpb := *pb
				clone.Publish[i] = &cpb
			}
		}
	}
	if spec.Mounts != nil {
		clone.Mounts = make([]*types.Mount, len(spec.Mounts))
		for i, m := range spec.Mounts {
			if m != nil {
				cm := *m
				clone.Mounts[i] = &cm
			}
		}
	}
	if spec.VolumesFrom != nil {
		clone.VolumesFrom = slices.Clone(spec.VolumesFrom)
	}
	if spec.StorageOpt != nil {
		clone.StorageOpt = maps.Clone(spec.StorageOpt)
	}
	if spec.GroupAdd != nil {
		clone.GroupAdd = slices.Clone(spec.GroupAdd)
	}
	if spec.CapAdd != nil {
		clone.CapAdd = slices.Clone(spec.CapAdd)
	}
	if spec.CapDrop != nil {
		clone.CapDrop = slices.Clone(spec.CapDrop)
	}
	if spec.SecurityOpt != nil {
		clone.SecurityOpt = slices.Clone(spec.SecurityOpt)
	}
	if spec.Sysctls != nil {
		clone.Sysctls = maps.Clone(spec.Sysctls)
	}
	return (*Template)(&clone)
}

func (t *Template) Clone() *Template {
	return t.Copy()
}

func (t *Template) ContainerSpec() *types.ContainerSpec {
	if t == nil {
		return nil
	}
	return (*types.ContainerSpec)(t.Copy())
}
