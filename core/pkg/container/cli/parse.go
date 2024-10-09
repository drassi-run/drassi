package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/opts"
	docker "github.com/docker/docker/api/types/container"
	"github.com/google/shlex"
	"github.com/spf13/pflag"
)

func Parse(opts string) (*container.ContainerSpec, error) {
	flags := pflag.NewFlagSet("docker create", pflag.ContinueOnError)
	copts := addFlags(flags)

	if args, err := shlex.Split(opts); err != nil {
		return nil, err
	} else if err = flags.Parse(args); err != nil {
		return nil, err
	}
	return parse(flags, copts)
}

// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L338-L740
func parse(flags *pflag.FlagSet, copts *containerOptions) (*container.ContainerSpec, error) {
	spec := container.ContainerSpec{
		Annotations: copts.annotations.GetAll(),
		Labels:      opts.ConvertKVStringsToMap(copts.labels.GetAll()), // TODO: label-files
		//ContainerName:
		//Image
		//PullPolicy
		//Command
		//Entrypoint
		Environment: opts.ConvertKVStringsToMapWithNil(copts.env.GetAll()),
		//EnvFiles:
		WorkingDir: copts.workingDir,
	}

	// Networking
	//TODO: networks
	if publishOpts := copts.publish.GetAll(); len(publishOpts) > 0 {
		for _, opt := range publishOpts {
			if conf, err := types.ParsePortConfig(opt); err != nil {
				return nil, err
			} else {
				spec.Ports = append(spec.Ports, conf...)
			}
		}
	}
	spec.Expose = copts.expose.GetAll()
	spec.DNS = copts.dns.GetAll()
	spec.DNSOpts = copts.dnsOptions.GetAll()
	spec.DNSSearch = copts.dnsSearch.GetAll()
	spec.DomainName = copts.domainname
	spec.Hostname = copts.hostname
	if extraHosts, err := types.NewHostsList(copts.extraHosts.GetAll()); err != nil {
		return nil, err
	} else {
		spec.ExtraHosts = extraHosts
	}

	// Storage & Device
	if mounts, err := parseMount(copts.mounts); err != nil {
		return nil, err
	} else if volumes, err := parseVolumes(copts.volumes); err != nil {
		return nil, err
	} else {
		spec.Volumes = append(mounts, volumes...)
	}
	spec.VolumeDriver = copts.volumeDriver
	spec.VolumesFrom = copts.volumesFrom.GetAll()
	spec.Tmpfs = copts.tmpfs.GetAll()
	if storageOpts, err := parseStorageOpts(copts.storageOpt.GetAll()); err != nil {
		return nil, err
	} else {
		spec.StorageOpt = storageOpts
	}
	spec.Devices = copts.devices.GetAll() // TODO https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L468-L493
	spec.DeviceCgroupRules = copts.deviceCgroupRules.GetAll()

	// Runtime
	spec.Runtime = copts.runtime
	//spec.Platform = copts.platform
	spec.Isolation = copts.isolation
	spec.StopSignal = copts.stopSignal
	if flags.Changed("stop-timeout") {
		d := time.Duration(copts.stopTimeout) * time.Second
		spec.StopGracePeriod = durationPtr(d)
	}

	if copts.loggingDriver != "" || copts.loggingOpts.Len() > 0 {
		loggingOpts, err := parseLoggingOpts(copts.loggingDriver, copts.loggingOpts.GetAll())
		if err != nil {
			return nil, err
		}
		spec.Logging = &types.LoggingConfig{
			Driver:  copts.loggingDriver,
			Options: loggingOpts,
		}
	}

	haveHealthSettings := copts.healthCmd != "" ||
		copts.healthInterval != 0 ||
		copts.healthTimeout != 0 ||
		copts.healthStartPeriod != 0 ||
		copts.healthRetries != 0 ||
		copts.healthStartInterval != 0
	if copts.noHealthcheck {
		if haveHealthSettings {
			return nil, fmt.Errorf("--no-healthcheck conflicts with --health-* options")
		}
	} else if haveHealthSettings {
		var probe []string
		if copts.healthCmd != "" {
			probe = []string{"CMD-SHELL", copts.healthCmd}
		}
		if copts.healthInterval < 0 {
			return nil, fmt.Errorf("--health-interval cannot be negative")
		}
		if copts.healthTimeout < 0 {
			return nil, fmt.Errorf("--health-timeout cannot be negative")
		}
		if copts.healthRetries < 0 {
			return nil, fmt.Errorf("--health-retries cannot be negative")
		}
		if copts.healthStartPeriod < 0 {
			return nil, fmt.Errorf("--health-start-period cannot be negative")
		}
		if copts.healthStartInterval < 0 {
			return nil, fmt.Errorf("--health-start-interval cannot be negative")
		}

		spec.HealthCheck = &types.HealthCheckConfig{
			Test:          probe,
			Interval:      durationPtr(copts.healthInterval),
			Timeout:       durationPtr(copts.healthTimeout),
			StartPeriod:   durationPtr(copts.healthStartPeriod),
			StartInterval: durationPtr(copts.healthStartInterval),
			Retries:       pointerOf(uint64(copts.healthRetries)),
		}
	}

	// Resources
	//// Applicable to all platforms
	spec.CPUShares = copts.cpuShares
	spec.CPUS = copts.cpus.String()
	spec.Memory = types.UnitBytes(copts.memory)
	//// Applicable to Windows
	spec.CPUCount = copts.cpuCount
	spec.CPUPercent = float32(copts.cpuPercent) / 100.0
	//// Applicable to UNIX
	spec.CPUPeriod = copts.cpuPeriod
	spec.CPUQuota = copts.cpuQuota
	spec.CPURTPeriod = copts.cpuRealtimePeriod
	spec.CPURTRuntime = copts.cpuRealtimeRuntime
	spec.CpusetCpus = copts.cpusetCpus
	spec.CpusetMems = copts.cpusetMems
	spec.MemReservation = types.UnitBytes(copts.memoryReservation)
	spec.MemSwapLimit = types.UnitBytes(copts.memorySwap)
	spec.MemSwappiness = types.UnitBytes(copts.swappiness)
	spec.ShmSize = types.UnitBytes(copts.shmSize)
	spec.OomKillDisable = copts.oomKillDisable
	spec.OomScoreAdj = int64(copts.oomScoreAdj)
	spec.PidsLimit = copts.pidsLimit

	spec.BlkioConfig = parseBlkioOpts(copts)
	if ulimits, err := validateUlimitsOpts(copts.ulimits); err != nil {
		return nil, err
	} else {
		spec.Ulimits = ulimits
	}

	// Namespace & CGroup
	networkMode := docker.NetworkMode(copts.netMode.NetworkMode())
	spec.NetworkMode = string(networkMode)

	if pidMode := docker.PidMode(copts.pidMode); !pidMode.Valid() {
		return nil, fmt.Errorf("--pid: invalid PID mode")
	} else {
		spec.PidMode = string(pidMode)
	}

	if utsMode := docker.UTSMode(copts.utsMode); !utsMode.Valid() {
		return nil, fmt.Errorf("--uts: invalid UTS mode")
	} else {
		spec.UTSMode = string(utsMode)
	}

	if usernsMode := docker.UsernsMode(copts.usernsMode); !usernsMode.Valid() {
		return nil, fmt.Errorf("--userns: invalid USER mode")
	} else {
		spec.UserMode = string(usernsMode)
	}

	if cgroupnsMode := docker.CgroupnsMode(copts.cgroupnsMode); !cgroupnsMode.Valid() {
		return nil, fmt.Errorf("--cgroupns: invalid CGROUP mode")
	} else {
		spec.CgroupMode = string(cgroupnsMode)
	}
	spec.CgroupParent = copts.cgroupParent

	// Security
	spec.User = copts.user
	spec.GroupAdd = copts.groupAdd.GetAll()
	spec.CapAdd = copts.capAdd.GetAll()
	spec.CapDrop = copts.capDrop.GetAll()
	spec.Privileged = copts.privileged
	spec.SecurityOpt = copts.securityOpt.GetAll()

	securityOpts, err := parseSecurityOpts(copts.securityOpt.GetAll())
	if err != nil {
		return nil, err
	}
	spec.SecurityOpt = securityOpts // TODO: parseSystemPaths https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L542
	spec.Sysctls = copts.sysctls.GetAll()

	return &spec, nil
}

func durationPtr(value time.Duration) *types.Duration {
	result := types.Duration(value)
	return &result
}

func pointerOf[V any](value V) *V {
	return &value
}

func parseVolumes(volumeOpt opts.ListOpts) ([]types.ServiceVolumeConfig, error) {
	volumes := make([]types.ServiceVolumeConfig, 0)
	for _, v := range volumeOpt.GetAll() {
		parsed, err := loader.ParseVolume(v)
		if err != nil {
			return nil, err
		}
		volume := types.ServiceVolumeConfig{
			Type:        parsed.Type,
			Source:      parsed.Source,
			Target:      parsed.Target,
			ReadOnly:    parsed.ReadOnly,
			Consistency: parsed.Consistency,
		}
		if parsed.Bind != nil {
			volume.Bind = &types.ServiceVolumeBind{
				Propagation: parsed.Bind.Propagation,
			}
		}
		if parsed.Volume != nil {
			volume.Volume = &types.ServiceVolumeVolume{
				NoCopy: parsed.Volume.NoCopy,
			}
		}
		if parsed.Tmpfs != nil {
			volume.Tmpfs = &types.ServiceVolumeTmpfs{
				Size: types.UnitBytes(parsed.Tmpfs.Size),
			}
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

func parseMount(mountOpt opts.MountOpt) ([]types.ServiceVolumeConfig, error) {
	volumes := make([]types.ServiceVolumeConfig, 0)
	for _, m := range mountOpt.Value() {
		volume := types.ServiceVolumeConfig{
			Type:        string(m.Type),
			Source:      m.Source,
			Target:      m.Target,
			ReadOnly:    m.ReadOnly,
			Consistency: string(m.Consistency),
		}
		if m.BindOptions != nil {
			volume.Bind = &types.ServiceVolumeBind{
				Propagation: string(m.BindOptions.Propagation),
				// TODO: other BindOptions options
			}
		}
		if m.VolumeOptions != nil {
			volume.Volume = &types.ServiceVolumeVolume{
				NoCopy:  m.VolumeOptions.NoCopy,
				Subpath: m.VolumeOptions.Subpath,
				// TODO: other VolumeOptions options
			}
		}
		if m.TmpfsOptions != nil {
			volume.Tmpfs = &types.ServiceVolumeTmpfs{
				Size: types.UnitBytes(m.TmpfsOptions.SizeBytes),
				Mode: uint32(m.TmpfsOptions.Mode),
			}
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
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
			Rate: types.UnitBytes(o.Rate),
		}
	}
	return td
}

// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L912-L918
func parseLoggingOpts(loggingDriver string, loggingOpts []string) (map[string]string, error) {
	loggingOptsMap := opts.ConvertKVStringsToMap(loggingOpts)
	if loggingDriver == "none" && len(loggingOpts) > 0 {
		return map[string]string{}, fmt.Errorf("invalid logging opts for driver %s", loggingDriver)
	}
	return loggingOptsMap, nil
}

// parses storage options per container into a map
// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L974-L984
func parseStorageOpts(storageOpts []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, option := range storageOpts {
		k, v, ok := strings.Cut(option, "=")
		if !ok {
			return nil, fmt.Errorf("invalid storage option")
		}
		m[k] = v
	}
	return m, nil
}

const (
	// seccompProfileDefault is the built-in default seccomp profile.
	seccompProfileDefault = "builtin"
	// seccompProfileUnconfined is a special profile name for seccomp to use an
	// "unconfined" seccomp profile.
	seccompProfileUnconfined = "unconfined"
)

// takes a local seccomp daemon, reads the file contents for sending to the daemon
// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L921-L952
func parseSecurityOpts(securityOpts []string) ([]string, error) {
	for key, opt := range securityOpts {
		k, v, ok := strings.Cut(opt, "=")
		if !ok && k != "no-new-privileges" {
			k, v, ok = strings.Cut(opt, ":")
		}
		if (!ok || v == "") && k != "no-new-privileges" {
			// "no-new-privileges" is the only option that does not require a value.
			return securityOpts, fmt.Errorf("invalid --security-opt: %q", opt)
		}
		if k == "seccomp" {
			switch v {
			case seccompProfileDefault, seccompProfileUnconfined:
				// known special names for built-in profiles, nothing to do.
			default:
				// value may be a filename, in which case we send the profile's
				// content if it's valid JSON.
				f, err := os.ReadFile(v)
				if err != nil {
					return securityOpts, fmt.Errorf("opening seccomp profile (%s) failed: %w", v, err)
				}
				b := bytes.NewBuffer(nil)
				if err := json.Compact(b, f); err != nil {
					return securityOpts, fmt.Errorf("compacting json for seccomp profile (%s) failed: %w", v, err)
				}
				securityOpts[key] = fmt.Sprintf("seccomp=%s", b.Bytes())
			}
		}
	}

	return securityOpts, nil
}
