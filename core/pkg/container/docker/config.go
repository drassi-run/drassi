package docker

import (
	"fmt"
	"time"

	"drassi.run/core/pkg/container/types"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	dockeropts "github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/blkiodev"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockernetwork "github.com/docker/docker/api/types/network"
)

type containerConfig struct {
	Config           *dockercontainer.Config
	HostConfig       *dockercontainer.HostConfig
	NetworkingConfig *dockernetwork.NetworkingConfig
}

func (cc *containerConfig) From(spec *types.ContainerSpec, stdio *types.Stdio) error {
	resources := dockercontainer.Resources{
		CgroupParent:      spec.CgroupParent,
		Memory:            spec.Memory,
		MemoryReservation: spec.MemReservation,
		MemorySwap:        spec.MemSwapLimit,
		MemorySwappiness:  pointerOf(spec.MemSwappiness),
		//KernelMemory:         spec.kernelMemory,
		OomKillDisable:     &spec.OomKillDisable,
		CPUCount:           spec.CPUCount,
		CPUPercent:         int64(spec.CPUPercent * 100),
		CPUShares:          spec.CPUShares,
		CPUPeriod:          spec.CPUPeriod,
		CpusetCpus:         spec.CpusetCpus,
		CpusetMems:         spec.CpusetMems,
		CPUQuota:           spec.CPUQuota,
		CPURealtimePeriod:  spec.CPURTPeriod,
		CPURealtimeRuntime: spec.CPURTRuntime,
		PidsLimit:          &spec.PidsLimit,
		IOMaximumIOps:      spec.IOMaximumIOps,
		IOMaximumBandwidth: spec.IOMaximumBandwidth,
		Ulimits:            spec.Ulimits,
		DeviceCgroupRules:  spec.DeviceCgroupRules,
		//Devices:           spec.Devices,
		//DeviceRequests:    deviceRequests,
	}
	if spec.CPUS != "" {
		if cpu, err := dockeropts.ParseCPUs(spec.CPUS); err != nil {
			return err
		} else {
			resources.NanoCPUs = cpu
		}
	}
	if blkio := spec.BlkioConfig; blkio != nil {
		resources.BlkioWeight = blkio.Weight
		resources.BlkioWeightDevice = cc.blkioWeightDeviceFrom(blkio.WeightDevice)
		resources.BlkioDeviceReadBps = cc.blkioThrottleDeviceFrom(blkio.DeviceReadBps)
		resources.BlkioDeviceWriteBps = cc.blkioThrottleDeviceFrom(blkio.DeviceWriteBps)
		resources.BlkioDeviceReadIOps = cc.blkioThrottleDeviceFrom(blkio.DeviceReadIOps)
		resources.BlkioDeviceWriteIOps = cc.blkioThrottleDeviceFrom(blkio.DeviceWriteIOps)
	}

	cc.Config = &dockercontainer.Config{
		User:         spec.User,
		Tty:          stdio.Tty,
		OpenStdin:    stdio.Interactive,
		AttachStdin:  stdio.AttachStdin(),
		AttachStdout: stdio.AttachStdout(),
		AttachStderr: stdio.AttachStderr(),
		Env:          convertMapToKVString(spec.Environment),
		Cmd:          spec.Command,
		Image:        spec.Image,
		Volumes:      nil,
		//MacAddress:   spec.MacAddress,
		Entrypoint:  spec.Entrypoint,
		WorkingDir:  spec.WorkingDir,
		Labels:      spec.Labels,
		StopSignal:  spec.StopSignal,
		StopTimeout: cc.stopTimeoutFrom(spec.StopGracePeriod),
	}
	if hc := spec.HealthCheck; hc != nil {
		cc.setHealCheck(spec.HealthCheck)
	}

	cc.HostConfig = &dockercontainer.HostConfig{
		ContainerIDFile: "",
		OomScoreAdj:     int(spec.OomScoreAdj),
		//AutoRemove:      spec.AutoRemove,
		Privileged:   spec.Privileged,
		IpcMode:      dockercontainer.IpcMode(spec.IpcMode),
		NetworkMode:  dockercontainer.NetworkMode(spec.NetworkMode),
		PidMode:      dockercontainer.PidMode(spec.PidMode),
		UTSMode:      dockercontainer.UTSMode(spec.UTSMode),
		UsernsMode:   dockercontainer.UsernsMode(spec.UserMode),
		CgroupnsMode: dockercontainer.CgroupnsMode(spec.CgroupMode),
		CapAdd:       spec.CapAdd,
		CapDrop:      spec.CapDrop,
		GroupAdd:     spec.GroupAdd,
		//RestartPolicy:  restartPolicy,
		SecurityOpt: spec.SecurityOpt,
		LogConfig:   cc.logConfigFrom(spec.Logging),
		Isolation:   dockercontainer.Isolation(spec.Isolation),
		ShmSize:     int64(spec.ShmSize),
		Resources:   resources,
		Sysctls:     spec.Sysctls,
		Runtime:     spec.Runtime,
		//MaskedPaths:   maskedPaths,
		Annotations: spec.Annotations,
	}

	cc.setStorage(&spec.ContainerStorage)
	cc.setNetwork(&spec.ContainerNetwork)

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

func (cc *containerConfig) logConfigFrom(lc *types.LoggingConfig) dockercontainer.LogConfig {
	if lc == nil {
		return dockercontainer.LogConfig{}
	}
	return dockercontainer.LogConfig{
		Type:   lc.Driver,
		Config: lc.Options,
	}
}

func (cc *containerConfig) setHealCheck(hc *types.HealthCheckConfig) {
	cc.Config.Healthcheck = &dockercontainer.HealthConfig{
		Test:          hc.Test,
		Timeout:       hc.Timeout,
		Interval:      hc.Interval,
		Retries:       hc.Retries,
		StartPeriod:   hc.StartPeriod,
		StartInterval: hc.StartInterval,
	}
}

func (cc *containerConfig) stopTimeoutFrom(d *composetypes.Duration) *int {
	if d == nil {
		return nil
	}
	s := int(time.Duration(*d).Seconds())
	return &s
}

func convertMapToKVString(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}

	r := make([]string, 0, len(m))
	for k, v := range m {
		r = append(r, fmt.Sprintf("%s=%s", k, v))
	}
	return r
}

func pointerOf[V any](value V) *V {
	return &value
}
