/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package docker

import (
	"drassi.run/core/pkg/container/types"
	dockeropts "github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/blkiodev"
	dockercontainer "github.com/docker/docker/api/types/container"
)

func (cc *containerConfig) setResources(conf *types.ContainerResource) error {
	res := dockercontainer.Resources{
		//// Applicable to all platforms
		CPUShares: conf.CPUShares,
		Memory:    conf.Memory,

		//// Applicable to Windows
		CPUCount:           conf.CPUCount,
		CPUPercent:         int64(conf.CPUPercent * 100),
		IOMaximumIOps:      conf.IOMaximumIOps,
		IOMaximumBandwidth: conf.IOMaximumBandwidth,

		//// Applicable to UNIX
		CPUPeriod:          conf.CPUPeriod,
		CPUQuota:           conf.CPUQuota,
		CPURealtimePeriod:  conf.CPURTPeriod,
		CPURealtimeRuntime: conf.CPURTRuntime,
		CpusetCpus:         conf.CpusetCpus,
		CpusetMems:         conf.CpusetMems,
		MemoryReservation:  conf.MemReservation,
		MemorySwap:         conf.MemSwapLimit,
		MemorySwappiness:   &conf.MemSwappiness,
		OomKillDisable:     &conf.OomKillDisable,
		PidsLimit:          &conf.PidsLimit,
		Ulimits:            conf.Ulimits,
	}
	if conf.CPUS != "" {
		if cpu, err := dockeropts.ParseCPUs(conf.CPUS); err != nil {
			return err
		} else {
			res.NanoCPUs = cpu
		}
	}

	if blkio := conf.BlkioConfig; blkio != nil {
		res.BlkioWeight = blkio.Weight
		res.BlkioWeightDevice = cc.blkioWeightDeviceFrom(blkio.WeightDevice)
		res.BlkioDeviceReadBps = cc.blkioThrottleDeviceFrom(blkio.DeviceReadBps)
		res.BlkioDeviceWriteBps = cc.blkioThrottleDeviceFrom(blkio.DeviceWriteBps)
		res.BlkioDeviceReadIOps = cc.blkioThrottleDeviceFrom(blkio.DeviceReadIOps)
		res.BlkioDeviceWriteIOps = cc.blkioThrottleDeviceFrom(blkio.DeviceWriteIOps)
	}

	hc := cc.HostConfig
	hc.Resources = res
	hc.ShmSize = conf.ShmSize
	hc.OomScoreAdj = int(conf.OomScoreAdj)

	return nil
}

func (cc *containerConfig) blkioWeightDeviceFrom(wd []types.WeightDevice) []*blkiodev.WeightDevice {
	if len(wd) == 0 {
		return nil
	}
	a := make([]*blkiodev.WeightDevice, len(wd))
	for i, w := range wd {
		a[i] = &blkiodev.WeightDevice{
			Path:   w.Path,
			Weight: w.Weight,
		}
	}
	return a
}

func (cc *containerConfig) blkioThrottleDeviceFrom(td []types.ThrottleDevice) []*blkiodev.ThrottleDevice {
	if len(td) == 0 {
		return nil
	}
	a := make([]*blkiodev.ThrottleDevice, len(td))
	for i, t := range td {
		a[i] = &blkiodev.ThrottleDevice{
			Path: t.Path,
			Rate: t.Rate,
		}
	}
	return a
}

func (cs *containerSpec) setResources(hc *dockercontainer.HostConfig) {
	res := &hc.Resources
	cpu := dockeropts.NanoCPUs(res.NanoCPUs)
	cs.Spec.ContainerResource = types.ContainerResource{
		//// Applicable to all platforms
		CPUShares: res.CPUShares,
		CPUS:      cpu.String(),
		Memory:    res.Memory,

		//// Applicable to Windows
		CPUCount:           res.CPUCount,
		CPUPercent:         float32(res.CPUPercent) / 100,
		IOMaximumIOps:      res.IOMaximumIOps,
		IOMaximumBandwidth: res.IOMaximumBandwidth,

		//// Applicable to UNIX
		CPUPeriod:      res.CPUPeriod,
		CPUQuota:       res.CPUQuota,
		CPURTPeriod:    res.CPURealtimePeriod,
		CPURTRuntime:   res.CPURealtimeRuntime,
		CpusetCpus:     res.CpusetCpus,
		CpusetMems:     res.CpusetMems,
		MemReservation: res.MemoryReservation,
		MemSwapLimit:   res.MemorySwap,
	}
	r := &cs.Spec.ContainerResource
	if res.MemorySwappiness != nil {
		r.MemSwappiness = *res.MemorySwappiness
	}
	if res.OomKillDisable != nil {
		r.OomKillDisable = *res.OomKillDisable
	}
	if res.PidsLimit != nil {
		r.PidsLimit = *res.PidsLimit
	}
	r.ShmSize = hc.ShmSize
	r.OomScoreAdj = int64(hc.OomScoreAdj)

	r.Ulimits = res.Ulimits

	if cs.anyBlkioConfig(res) {
		r.BlkioConfig = &types.BlkioConfig{
			Weight:          res.BlkioWeight,
			WeightDevice:    cs.blkioWeightDeviceFrom(res.BlkioWeightDevice),
			DeviceReadBps:   cs.blkioThrottleDeviceFrom(res.BlkioDeviceReadBps),
			DeviceWriteBps:  cs.blkioThrottleDeviceFrom(res.BlkioDeviceWriteBps),
			DeviceReadIOps:  cs.blkioThrottleDeviceFrom(res.BlkioDeviceReadIOps),
			DeviceWriteIOps: cs.blkioThrottleDeviceFrom(res.BlkioDeviceWriteIOps),
		}
	}
}

func (cs *containerSpec) anyBlkioConfig(res *dockercontainer.Resources) bool {
	if res.BlkioWeight != 0 {
		return true
	}
	if len(res.BlkioWeightDevice) != 0 {
		return true
	}
	if len(res.BlkioDeviceReadBps) != 0 {
		return true
	}
	if len(res.BlkioDeviceWriteBps) != 0 {
		return true
	}
	if len(res.BlkioDeviceReadIOps) != 0 {
		return true
	}
	if len(res.BlkioDeviceWriteIOps) != 0 {
		return true
	}
	return false
}

func (cs *containerSpec) blkioWeightDeviceFrom(wd []*blkiodev.WeightDevice) []types.WeightDevice {
	if len(wd) == 0 {
		return nil
	}
	a := make([]types.WeightDevice, len(wd))
	for i, w := range wd {
		a[i] = types.WeightDevice{
			Path:   w.Path,
			Weight: w.Weight,
		}
	}
	return a
}

func (cs *containerSpec) blkioThrottleDeviceFrom(td []*blkiodev.ThrottleDevice) []types.ThrottleDevice {
	if len(td) == 0 {
		return nil
	}
	a := make([]types.ThrottleDevice, len(td))
	for i, t := range td {
		a[i] = types.ThrottleDevice{
			Path: t.Path,
			Rate: t.Rate,
		}
	}
	return a
}
