package cli

import (
	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/opts"
)

func (fm *flagMapper) mapResource(copts *containerOptions) error {
	res := &fm.Spec.ContainerResource

	// Resources
	//// Applicable to all platforms
	res.CPUShares = copts.cpuShares
	res.CPUS = copts.cpus.String()
	res.Memory = int64(copts.memory)
	//// Applicable to Windows
	res.CPUCount = copts.cpuCount
	res.CPUPercent = float32(copts.cpuPercent) / 100.0
	//// Applicable to UNIX
	res.CPUPeriod = copts.cpuPeriod
	res.CPUQuota = copts.cpuQuota
	res.CPURTPeriod = copts.cpuRealtimePeriod
	res.CPURTRuntime = copts.cpuRealtimeRuntime
	res.CpusetCpus = copts.cpusetCpus
	res.CpusetMems = copts.cpusetMems
	res.MemReservation = int64(copts.memoryReservation)
	res.MemSwapLimit = int64(copts.memorySwap)
	res.MemSwappiness = copts.swappiness
	res.ShmSize = int64(copts.shmSize)
	res.OomKillDisable = copts.oomKillDisable
	res.OomScoreAdj = int64(copts.oomScoreAdj)
	res.PidsLimit = copts.pidsLimit

	res.BlkioConfig = parseBlkioOpts(copts)
	res.Ulimits = copts.ulimits.GetList()
	return nil
}

func parseBlkioOpts(copts *containerOptions) *types.BlkioConfig {
	if copts.blkioWeight == 0 &&
		len(copts.blkioWeightDevice.GetList()) == 0 &&
		len(copts.deviceReadBps.GetList()) == 0 &&
		len(copts.deviceReadIOps.GetList()) == 0 &&
		len(copts.deviceWriteBps.GetList()) == 0 &&
		len(copts.deviceWriteIOps.GetList()) == 0 {
		return nil
	}
	weightDevice := parseWeightDeviceOpts(copts.blkioWeightDevice)
	deviceReadBps := parseThrottleDeviceOpts(copts.deviceReadBps)
	deviceReadIOps := parseThrottleDeviceOpts(copts.deviceReadIOps)
	deviceWriteBps := parseThrottleDeviceOpts(copts.deviceWriteBps)
	deviceWriteIOps := parseThrottleDeviceOpts(copts.deviceWriteIOps)
	return &types.BlkioConfig{
		Weight:          copts.blkioWeight,
		WeightDevice:    weightDevice,
		DeviceReadBps:   deviceReadBps,
		DeviceReadIOps:  deviceReadIOps,
		DeviceWriteBps:  deviceWriteBps,
		DeviceWriteIOps: deviceWriteIOps,
	}
}

func parseWeightDeviceOpts(opt opts.WeightdeviceOpt) []types.WeightDevice {
	if len(opt.GetList()) <= 0 {
		return nil
	}
	wd := make([]types.WeightDevice, len(opt.GetList()))
	for i, o := range opt.GetList() {
		wd[i] = types.WeightDevice{
			Path:   o.Path,
			Weight: o.Weight,
		}
	}
	return wd
}

func parseThrottleDeviceOpts(opt opts.ThrottledeviceOpt) []types.ThrottleDevice {
	if len(opt.GetList()) <= 0 {
		return nil
	}
	td := make([]types.ThrottleDevice, len(opt.GetList()))
	for i, o := range opt.GetList() {
		td[i] = types.ThrottleDevice{
			Path: o.Path,
			Rate: o.Rate,
		}
	}
	return td
}
