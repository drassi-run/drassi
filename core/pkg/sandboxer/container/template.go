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
)

const DefaultImage = "ghcr.io/drassi-run/ubuntu:26.04"

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
	s := *t
	s.Command = slices.Clone(t.Command)
	s.Entrypoint = slices.Clone(t.Entrypoint)
	s.Environment = maps.Clone(t.Environment)
	s.Labels = maps.Clone(t.Labels)
	s.Annotations = maps.Clone(t.Annotations)
	s.Devices = slices.Clone(t.Devices)
	s.DeviceCgroupRules = slices.Clone(t.DeviceCgroupRules)
	s.Exposes = clone(t.Exposes)
	s.Publish = clone(t.Publish)
	s.Mounts = clone(t.Mounts)
	s.VolumesFrom = slices.Clone(t.VolumesFrom)
	s.StorageOpt = maps.Clone(t.StorageOpt)
	s.GroupAdd = slices.Clone(t.GroupAdd)
	s.CapAdd = slices.Clone(t.CapAdd)
	s.CapDrop = slices.Clone(t.CapDrop)
	s.SecurityOpt = slices.Clone(t.SecurityOpt)
	s.Sysctls = maps.Clone(t.Sysctls)
	return &s
}

func (t *Template) ContainerSpec() *types.ContainerSpec {
	if t == nil {
		return nil
	}
	return (*types.ContainerSpec)(t.Copy())
}

func clone[S ~[]*E, E any](s S) S {
	if s == nil {
		return nil
	}

	s2 := make(S, len(s))
	for idx, i := range s {
		if i != nil {
			j := *i
			s2[idx] = &j
		}
	}
	return s2
}
