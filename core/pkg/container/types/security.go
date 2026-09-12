/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

type ContainerSecurity struct {
	// Namespace & CGroup
	NetworkMode  string `json:"network_mode,omitempty"`
	IpcMode      string `json:"ipc_mode,omitempty"`
	PidMode      string `json:"pid_mode,omitempty"`
	UTSMode      string `json:"uts_mode,omitempty"`
	UserMode     string `json:"user_mode,omitempty"`
	CgroupMode   string `json:"cgroup_mode,omitempty"`
	CgroupParent string `json:"cgroup_parent,omitempty"`

	// Security
	User        string   `json:"user,omitempty"`
	GroupAdd    []string `json:"group_add,omitempty"`
	CapAdd      []string `json:"cap_add,omitempty"`
	CapDrop     []string `json:"cap_drop,omitempty"`
	Privileged  bool     `json:"privileged,omitempty"`
	SecurityOpt []string `json:"security_opt,omitempty"`
	Sysctls     Mapping  `json:"sysctls,omitempty"`
}
