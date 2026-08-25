/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"drassi.run/core/pkg/container/types"
	dockercontainer "github.com/moby/moby/api/types/container"
)

func (cc *containerConfig) setSecurity(conf *types.ContainerSecurity) {
	c, hc := cc.Config, cc.HostConfig

	hc.Privileged = conf.Privileged
	hc.NetworkMode = dockercontainer.NetworkMode(conf.NetworkMode)
	hc.IpcMode = dockercontainer.IpcMode(conf.IpcMode)
	hc.PidMode = dockercontainer.PidMode(conf.PidMode)
	hc.UTSMode = dockercontainer.UTSMode(conf.UTSMode)
	hc.UsernsMode = dockercontainer.UsernsMode(conf.UserMode)
	hc.CgroupnsMode = dockercontainer.CgroupnsMode(conf.CgroupMode)
	hc.CgroupParent = conf.CgroupParent

	c.User = conf.User
	hc.GroupAdd = conf.GroupAdd
	hc.CapAdd = conf.CapAdd
	hc.CapDrop = conf.CapDrop
	hc.SecurityOpt = conf.SecurityOpt
	hc.Sysctls = conf.Sysctls
}

func (cs *containerSpec) setSecurity(c *dockercontainer.Config, hc *dockercontainer.HostConfig) {
	cs.Spec.ContainerSecurity = types.ContainerSecurity{
		Privileged:   hc.Privileged,
		NetworkMode:  string(hc.NetworkMode),
		IpcMode:      string(hc.IpcMode),
		PidMode:      string(hc.PidMode),
		UTSMode:      string(hc.UTSMode),
		UserMode:     string(hc.UsernsMode),
		CgroupMode:   string(hc.CgroupnsMode),
		CgroupParent: hc.CgroupParent,

		User:        c.User,
		GroupAdd:    hc.GroupAdd,
		CapAdd:      hc.CapAdd,
		CapDrop:     hc.CapDrop,
		SecurityOpt: hc.SecurityOpt,
		Sysctls:     hc.Sysctls,
	}
}
