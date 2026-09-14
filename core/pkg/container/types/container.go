/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

type ContainerSpec struct {
	Name       string `json:"name,omitempty"`
	Image      string `json:"image,omitempty"`
	PullPolicy string `json:"pull_policy,omitempty"`

	Command     []string `json:"command,omitempty"`
	Entrypoint  []string `json:"entrypoint,omitempty"`
	WorkingDir  string   `json:"working_dir,omitempty"`
	Environment Mapping  `json:"environment,omitempty"`
	Labels      Mapping  `json:"labels,omitempty"`
	Annotations Mapping  `json:"annotations,omitempty"`

	ContainerNetwork  `json:",embed"`
	ContainerStorage  `json:",embed"`
	Devices           []string `json:"devices,omitempty"`
	DeviceCgroupRules []string `json:"device_cgroup_rules,omitempty"`

	ContainerRuntime  `json:",embed"`
	ContainerResource `json:",embed"`
	ContainerSecurity `json:",embed"`
}
