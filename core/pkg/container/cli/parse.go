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

func Parse(opts string) (*container.ContainerSpec, *container.Stdio, error) {
	flags := pflag.NewFlagSet("docker create", pflag.ContinueOnError)
	copts := addFlags(flags)

	if args, err := shlex.Split(opts); err != nil {
		return nil, nil, err
	} else if err = flags.Parse(args); err != nil {
		return nil, nil, err
	}

	fm := new(flagMapper)
	if err := fm.Map(flags, copts); err != nil {
		return nil, nil, err
	}

	return &fm.Spec, &fm.Stdio, nil
}

type flagMapper struct {
	Spec  container.ContainerSpec
	Stdio container.Stdio
}

// Map function convert flags and copts to Spec and Stdio
// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L338-L740
func (fm *flagMapper) Map(flags *pflag.FlagSet, copts *containerOptions) error {
	fm.mapStdio(copts)

	fm.Spec.Name = copts.name
	fm.Spec.Labels = opts.ConvertKVStringsToMap(copts.labels.GetAll())
	fm.Spec.Annotations = copts.annotations.GetAll()
	fm.Spec.PullPolicy = copts.pull
	fm.Spec.Environment = opts.ConvertKVStringsToMap(copts.env.GetAll())
	fm.Spec.WorkingDir = copts.workingDir

	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L419-L424
	if copts.entrypoint != "" {
		fm.Spec.Entrypoint = []string{copts.entrypoint}
	} else if flags.Changed("entrypoint") {
		// if `--entrypoint=` is parsed then Entrypoint is reset
		fm.Spec.Entrypoint = []string{""}
	}

	// Networking
	//TODO: networks
	if publishOpts := copts.publish.GetAll(); len(publishOpts) > 0 {
		for _, opt := range publishOpts {
			if conf, err := types.ParsePortConfig(opt); err != nil {
				return err
			} else {
				fm.Spec.Ports = append(fm.Spec.Ports, conf...)
			}
		}
	}
	fm.Spec.Expose = copts.expose.GetAll()
	fm.Spec.DNS = copts.dns.GetAll()
	fm.Spec.DNSOpts = copts.dnsOptions.GetAll()
	fm.Spec.DNSSearch = copts.dnsSearch.GetAll()
	fm.Spec.DomainName = copts.domainname
	fm.Spec.Hostname = copts.hostname
	if extraHosts, err := types.NewHostsList(copts.extraHosts.GetAll()); err != nil {
		return err
	} else {
		fm.Spec.ExtraHosts = extraHosts
	}

	// Storage & Device
	if mounts, err := parseMount(copts.mounts); err != nil {
		return err
	} else if volumes, err := parseVolumes(copts.volumes); err != nil {
		return err
	} else {
		fm.Spec.Volumes = append(mounts, volumes...)
	}
	fm.Spec.VolumeDriver = copts.volumeDriver
	fm.Spec.VolumesFrom = copts.volumesFrom.GetAll()
	fm.Spec.Tmpfs = copts.tmpfs.GetAll()
	if storageOpts, err := parseStorageOpts(copts.storageOpt.GetAll()); err != nil {
		return err
	} else {
		fm.Spec.StorageOpt = storageOpts
	}
	fm.Spec.Devices = copts.devices.GetAll() // TODO https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L468-L493
	fm.Spec.DeviceCgroupRules = copts.deviceCgroupRules.GetAll()

	// Runtime
	fm.Spec.Runtime = copts.runtime
	//fm.Spec.Platform = copts.platform
	fm.Spec.Isolation = copts.isolation
	fm.Spec.StopSignal = copts.stopSignal
	if flags.Changed("stop-timeout") {
		d := time.Duration(copts.stopTimeout) * time.Second
		fm.Spec.StopGracePeriod = durationPtr(d)
	}

	if err := fm.mapLogging(copts); err != nil {
		return err
	}
	if err := fm.mapHealth(copts); err != nil {
		return err
	}

	// Resources
	//// Applicable to all platforms
	fm.Spec.CPUShares = copts.cpuShares
	fm.Spec.CPUS = copts.cpus.String()
	fm.Spec.Memory = types.UnitBytes(copts.memory)
	//// Applicable to Windows
	fm.Spec.CPUCount = copts.cpuCount
	fm.Spec.CPUPercent = float32(copts.cpuPercent) / 100.0
	//// Applicable to UNIX
	fm.Spec.CPUPeriod = copts.cpuPeriod
	fm.Spec.CPUQuota = copts.cpuQuota
	fm.Spec.CPURTPeriod = copts.cpuRealtimePeriod
	fm.Spec.CPURTRuntime = copts.cpuRealtimeRuntime
	fm.Spec.CpusetCpus = copts.cpusetCpus
	fm.Spec.CpusetMems = copts.cpusetMems
	fm.Spec.MemReservation = types.UnitBytes(copts.memoryReservation)
	fm.Spec.MemSwapLimit = types.UnitBytes(copts.memorySwap)
	fm.Spec.MemSwappiness = types.UnitBytes(copts.swappiness)
	fm.Spec.ShmSize = types.UnitBytes(copts.shmSize)
	fm.Spec.OomKillDisable = copts.oomKillDisable
	fm.Spec.OomScoreAdj = int64(copts.oomScoreAdj)
	fm.Spec.PidsLimit = copts.pidsLimit

	fm.Spec.BlkioConfig = parseBlkioOpts(copts)
	if ulimits, err := validateUlimitsOpts(copts.ulimits); err != nil {
		return err
	} else {
		fm.Spec.Ulimits = ulimits
	}

	// Namespace & CGroup
	networkMode := docker.NetworkMode(copts.netMode.NetworkMode())
	fm.Spec.NetworkMode = string(networkMode)

	if pidMode := docker.PidMode(copts.pidMode); !pidMode.Valid() {
		return fmt.Errorf("--pid: invalid PID mode")
	} else {
		fm.Spec.PidMode = string(pidMode)
	}

	if utsMode := docker.UTSMode(copts.utsMode); !utsMode.Valid() {
		return fmt.Errorf("--uts: invalid UTS mode")
	} else {
		fm.Spec.UTSMode = string(utsMode)
	}

	if usernsMode := docker.UsernsMode(copts.usernsMode); !usernsMode.Valid() {
		return fmt.Errorf("--userns: invalid USER mode")
	} else {
		fm.Spec.UserMode = string(usernsMode)
	}

	if cgroupnsMode := docker.CgroupnsMode(copts.cgroupnsMode); !cgroupnsMode.Valid() {
		return fmt.Errorf("--cgroupns: invalid CGROUP mode")
	} else {
		fm.Spec.CgroupMode = string(cgroupnsMode)
	}
	fm.Spec.CgroupParent = copts.cgroupParent

	// Security
	fm.Spec.User = copts.user
	fm.Spec.GroupAdd = copts.groupAdd.GetAll()
	fm.Spec.CapAdd = copts.capAdd.GetAll()
	fm.Spec.CapDrop = copts.capDrop.GetAll()
	fm.Spec.Privileged = copts.privileged
	fm.Spec.SecurityOpt = copts.securityOpt.GetAll()

	securityOpts, err := parseSecurityOpts(copts.securityOpt.GetAll())
	if err != nil {
		return err
	}
	fm.Spec.SecurityOpt = securityOpts // TODO: parseSystemPaths https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L542
	fm.Spec.Sysctls = copts.sysctls.GetAll()

	return nil
}

func (fm *flagMapper) mapStdio(copts *containerOptions) {
	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L339-L358
	fm.Stdio.Tty = copts.tty
	fm.Stdio.Interactive = copts.stdin // --interactive
	if copts.attach.Get("stdin") || copts.stdin {
		fm.Stdio.Attach |= container.Stdin
	}
	if copts.attach.Get("stdout") {
		fm.Stdio.Attach |= container.Stdout
	}
	if copts.attach.Get("stderr") {
		fm.Stdio.Attach |= container.Stderr
	}
	// If -a is not set, attach to stdout and stderr
	if copts.attach.Len() == 0 {
		fm.Stdio.Attach |= container.Stdout | container.Stderr
	}
}

func (fm *flagMapper) mapStorage(copts *containerOptions) error {
	//if mounts, err := parseMount(copts.mounts); err != nil {
	//	return err
	//} else if volumes, err := parseVolumes(copts.volumes); err != nil {
	//	return err
	//} else {
	//	fm.Spec.Volumes = append(mounts, volumes...)
	//}
	//fm.Spec.VolumeDriver = copts.volumeDriver
	//fm.Spec.VolumesFrom = copts.volumesFrom.GetAll()
	//fm.Spec.Tmpfs = copts.tmpfs.GetAll()
	//fm.Spec.ReadonlyRootfs = copts.readonlyRootfs
	//if storageOpts, err := parseStorageOpts(copts.storageOpt.GetAll()); err != nil {
	//	return nil, err
	//} else {
	//	fm.Spec.StorageOpt = storageOpts
	//}
	return nil
}

func (fm *flagMapper) mapLogging(copts *containerOptions) error {
	if copts.loggingDriver == "" && copts.loggingOpts.Len() == 0 {
		return nil
	}

	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L912-L918
	driver := copts.loggingDriver
	options := opts.ConvertKVStringsToMap(copts.loggingOpts.GetAll())
	if driver == "none" && len(options) > 0 {
		return fmt.Errorf("invalid logging opts for driver %s", driver)
	}

	fm.Spec.Logging = &types.LoggingConfig{
		Driver:  driver,
		Options: options,
	}
	return nil
}

func (fm *flagMapper) mapHealth(copts *containerOptions) error {
	haveHealthSettings := copts.healthCmd != "" ||
		copts.healthInterval != 0 ||
		copts.healthTimeout != 0 ||
		copts.healthStartPeriod != 0 ||
		copts.healthRetries != 0 ||
		copts.healthStartInterval != 0

	if !haveHealthSettings {
		return nil
	}

	if copts.noHealthcheck {
		return fmt.Errorf("--no-healthcheck conflicts with --health-* options")
	}

	var probe []string
	if copts.healthCmd != "" {
		probe = []string{"CMD-SHELL", copts.healthCmd}
	}
	if copts.healthInterval < 0 {
		return fmt.Errorf("--health-interval cannot be negative")
	}
	if copts.healthTimeout < 0 {
		return fmt.Errorf("--health-timeout cannot be negative")
	}
	if copts.healthRetries < 0 {
		return fmt.Errorf("--health-retries cannot be negative")
	}
	if copts.healthStartPeriod < 0 {
		return fmt.Errorf("--health-start-period cannot be negative")
	}
	if copts.healthStartInterval < 0 {
		return fmt.Errorf("--health-start-interval cannot be negative")
	}

	fm.Spec.HealthCheck = &container.HealthCheckConfig{
		Test:          probe,
		Timeout:       copts.healthTimeout,
		Interval:      copts.healthInterval,
		Retries:       copts.healthRetries,
		StartPeriod:   copts.healthStartPeriod,
		StartInterval: copts.healthStartInterval,
	}

	return nil
}

func durationPtr(value time.Duration) *types.Duration {
	result := types.Duration(value)
	return &result
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
