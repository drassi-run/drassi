package docker

import (
	"fmt"
	"io/fs"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
	"github.com/compose-spec/compose-go/v2/types"
	dockeropts "github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/blkiodev"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockermount "github.com/docker/docker/api/types/mount"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

type containerConfig struct {
	Config           *dockercontainer.Config
	HostConfig       *dockercontainer.HostConfig
	NetworkingConfig *dockernetwork.NetworkingConfig
}

func (cc *containerConfig) From(spec *container.ContainerSpec, stdio *container.Stdio) error {
	resources := dockercontainer.Resources{
		CgroupParent:      spec.CgroupParent,
		Memory:            int64(spec.Memory),
		MemoryReservation: int64(spec.MemReservation),
		MemorySwap:        int64(spec.MemSwapLimit),
		MemorySwappiness:  pointerOf(int64(spec.MemSwappiness)),
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
		//IOMaximumIOps:      spec.ioMaxIOps,
		//IOMaximumBandwidth: uint64(spec.ioMaxBandwidth),
		Ulimits:           cc.ulimitsFrom(spec.Ulimits),
		DeviceCgroupRules: spec.DeviceCgroupRules,
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
		Hostname:     spec.Hostname,
		Domainname:   spec.DomainName,
		ExposedPorts: cc.exposedPortsFrom(spec.Ports),
		User:         spec.User,
		Tty:          stdio.Tty,
		OpenStdin:    stdio.Interactive,
		AttachStdin:  stdio.AttachStdin(),
		AttachStdout: stdio.AttachStdout(),
		AttachStderr: stdio.AttachStderr(),
		Env:          convertMapToKVString(spec.Environment),
		Cmd:          spec.Command,
		Image:        spec.Image,
		Volumes:      cc.volumeFrom(spec.Volumes),
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
		Binds:           cc.bindFrom(spec.Volumes),
		ContainerIDFile: "",
		OomScoreAdj:     int(spec.OomScoreAdj),
		//AutoRemove:      spec.AutoRemove,
		Privileged:   spec.Privileged,
		PortBindings: cc.portBindingsFrom(spec.Ports),
		//Links:           spec.links.GetAll(),
		//PublishAllPorts: spec.publishAll,
		DNS:        spec.DNS,
		DNSSearch:  spec.DNSSearch,
		DNSOptions: spec.DNSOpts,
		ExtraHosts: spec.ExtraHosts.AsList("="),
		//VolumesFrom:    spec.volumesFrom.GetAll(),
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
		StorageOpt:  spec.StorageOpt,
		//ReadonlyRootfs: spec.readonlyRootfs,
		LogConfig:    cc.logConfigFrom(spec.Logging),
		VolumeDriver: spec.VolumeDriver,
		Isolation:    dockercontainer.Isolation(spec.Isolation),
		ShmSize:      int64(spec.ShmSize),
		Resources:    resources,
		Tmpfs:        cc.tmpfsFrom(spec.Volumes),
		Sysctls:      spec.Sysctls,
		Runtime:      spec.Runtime,
		Mounts:       cc.mountFrom(spec.Volumes),
		//MaskedPaths:   maskedPaths,
		//ReadonlyPaths: readonlyPaths,
		Annotations: spec.Annotations,
	}
	cc.NetworkingConfig = &dockernetwork.NetworkingConfig{
		EndpointsConfig: cc.networkEndpointsFrom(spec.Networks),
	}
	return nil
}

func (cc *containerConfig) networkEndpointsFrom(m map[string]*types.ServiceNetworkConfig) map[string]*dockernetwork.EndpointSettings {
	if len(m) == 0 {
		return nil
	}
	networks := make(map[string]*dockernetwork.EndpointSettings, len(m))
	for k, v := range m {
		if v == nil {
			networks[k] = nil
			continue
		}
		ep := dockernetwork.EndpointSettings{
			Aliases:    v.Aliases,
			MacAddress: v.MacAddress,
			DriverOpts: v.DriverOpts,
		}
		if v.Ipv4Address != "" || v.Ipv6Address != "" || len(v.LinkLocalIPs) > 0 {
			ep.IPAMConfig = &dockernetwork.EndpointIPAMConfig{
				IPv4Address:  v.Ipv4Address,
				IPv6Address:  v.Ipv6Address,
				LinkLocalIPs: v.LinkLocalIPs,
			}
		}
		networks[k] = &ep
	}
	return networks
}

func (cc *containerConfig) exposedPortsFrom(ports []types.ServicePortConfig) nat.PortSet {
	return nil
}

func (cc *containerConfig) portBindingsFrom(ports []types.ServicePortConfig) nat.PortMap {
	return nil
}

func (cc *containerConfig) bindFrom(volumes []types.ServiceVolumeConfig) []string {
	m := make([]string, 0)
	for _, v := range volumes {
		if v.Type == string(dockermount.TypeBind) {
			b := v.Source + ":" + v.Target
			m = append(m, b)
		}
	}
	return m
}

func (cc *containerConfig) volumeFrom(volumes []types.ServiceVolumeConfig) map[string]empty {
	m := make(map[string]struct{})
	for _, v := range volumes {
		if v.Type == string(dockermount.TypeVolume) {
			m[v.Source] = empty{}
		}
	}
	return m
}

func (cc *containerConfig) tmpfsFrom(volumes []types.ServiceVolumeConfig) map[string]string {
	m := make(map[string]string)
	for _, v := range volumes {
		if v.Type == string(dockermount.TypeTmpfs) {
			var opt []string
			if v.ReadOnly {
				opt = append(opt, "ro")
			}
			if c := v.Tmpfs; c != nil {
				if size := c.Size; size != 0 {
					opt = append(opt, fmt.Sprintf("size=%d", size))
				}
				if mode := c.Mode; mode != 0 {
					opt = append(opt, fmt.Sprintf("mode=%d", mode))
				}
			}
			m[v.Target] = strings.Join(opt, ",")
		}
	}
	return m
}

func (cc *containerConfig) mountFrom(volumes []types.ServiceVolumeConfig) []dockermount.Mount {
	mounts := make([]dockermount.Mount, len(volumes))
	for i, v := range volumes {
		m := dockermount.Mount{
			Type:        dockermount.Type(v.Type),
			Source:      v.Source,
			Target:      v.Target,
			ReadOnly:    v.ReadOnly,
			Consistency: dockermount.Consistency(v.Consistency),
		}
		if bind := v.Bind; bind != nil {
			m.BindOptions = &dockermount.BindOptions{
				Propagation: dockermount.Propagation(bind.Propagation),
			}
		}
		if volume := v.Volume; volume != nil {
			m.VolumeOptions = &dockermount.VolumeOptions{
				NoCopy:  volume.NoCopy,
				Subpath: volume.Subpath,
			}
		}
		if tmpfs := v.Tmpfs; tmpfs != nil {
			m.TmpfsOptions = &dockermount.TmpfsOptions{
				SizeBytes: int64(tmpfs.Size),
				Mode:      fs.FileMode(tmpfs.Mode),
			}
		}
		mounts[i] = m
	}
	return mounts
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
			Rate: uint64(t.Rate),
		}
	}
	return a
}

func (cc *containerConfig) ulimitsFrom(m map[string]types.UlimitsConfig) []*dockercontainer.Ulimit {
	if len(m) == 0 {
		return nil
	}
	ulimits := make([]*dockercontainer.Ulimit, 0, len(m))
	for k, v := range m {
		ulimits = append(ulimits, &dockercontainer.Ulimit{
			Name: k,
			Soft: int64(v.Soft),
			Hard: int64(v.Hard),
		})
	}
	return ulimits
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

func (cc *containerConfig) setHealCheck(hc *container.HealthCheckConfig) {
	cc.Config.Healthcheck = &dockercontainer.HealthConfig{
		Test:          hc.Test,
		Timeout:       hc.Timeout,
		Interval:      hc.Interval,
		Retries:       hc.Retries,
		StartPeriod:   hc.StartPeriod,
		StartInterval: hc.StartInterval,
	}
}

func (cc *containerConfig) stopTimeoutFrom(d *types.Duration) *int {
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

func nilToZero[V any](v *V) V {
	if v == nil {
		v = new(V)
	}
	return *v
}

func pointerOf[V any](value V) *V {
	return &value
}
