package cli

import (
	"fmt"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types/container"
	"github.com/google/shlex"
	"github.com/spf13/pflag"
)

func Parse(opts string) (*ContainerSpec, error) {
	flags := pflag.NewFlagSet("docker create", pflag.ContinueOnError)
	copts := addFlags(flags)

	if args, err := shlex.Split(opts); err != nil {
		return nil, err
	} else if err = flags.Parse(args); err != nil {
		return nil, err
	}
	return parse(flags, copts)
}

func parse(flags *pflag.FlagSet, copts *containerOptions) (*ContainerSpec, error) {
	config := ContainerSpec{
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
	//TODO: ports
	//TODO: expose
	config.DNS = copts.dns.GetAll()
	config.DNSOpts = copts.dnsOptions.GetAll()
	config.DNSSearch = copts.dnsSearch.GetAll()
	config.DomainName = copts.domainname
	config.Hostname = copts.hostname
	extraHosts, err := types.NewHostsList(copts.extraHosts.GetAll())
	if err != nil {
		return nil, err
	}
	config.ExtraHosts = extraHosts

	// Storage & Device
	mounts, err := parseMount(copts.mounts)
	if err != nil {
		return nil, err
	}
	volumes, err := parseVolumes(copts.volumes)
	if err != nil {
		return nil, err
	}
	config.Volumes = append(mounts, volumes...)
	config.VolumeDriver = copts.volumeDriver
	config.VolumesFrom = copts.volumesFrom.GetAll()
	config.Tmpfs = copts.tmpfs.GetAll()
	storageOpts, err := parseStorageOpts(copts.storageOpt.GetAll())
	if err != nil {
		return nil, err
	}
	config.StorageOpt = storageOpts
	config.Devices = copts.devices.GetAll() // TODO https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L468-L493
	config.DeviceCgroupRules = copts.deviceCgroupRules.GetAll()

	// Runtime
	config.Runtime = copts.runtime
	//config.Platform = copts.platform
	config.Isolation = copts.isolation
	config.StopSignal = copts.stopSignal
	if flags.Changed("stop-timeout") {
		d := time.Duration(copts.stopTimeout) * time.Second
		config.StopGracePeriod = durationPtr(d)
	}

	if copts.loggingDriver != "" || copts.loggingOpts.Len() > 0 {
		loggingOpts, err := parseLoggingOpts(copts.loggingDriver, copts.loggingOpts.GetAll())
		if err != nil {
			return nil, err
		}
		config.Logging = &types.LoggingConfig{
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

		config.HealthCheck = &types.HealthCheckConfig{
			Test:          probe,
			Interval:      durationPtr(copts.healthInterval),
			Timeout:       durationPtr(copts.healthTimeout),
			StartPeriod:   durationPtr(copts.healthStartPeriod),
			StartInterval: durationPtr(copts.healthStartInterval),
			Retries:       uint64Ptr(uint64(copts.healthRetries)),
		}
	}

	// Resources
	//// Applicable to all platforms
	config.CPUShares = copts.cpuShares
	config.CPUS = copts.cpus.String()
	config.Memory = types.UnitBytes(copts.memory.Value())
	//// Applicable to Windows
	config.CPUCount = copts.cpuCount
	config.CPUPercent = float32(copts.cpuPercent) / 100.0
	//// Applicable to UNIX
	config.CPUPeriod = copts.cpuPeriod
	config.CPUQuota = copts.cpuQuota
	config.CPURTPeriod = copts.cpuRealtimePeriod
	config.CPURTRuntime = copts.cpuRealtimeRuntime
	config.CpusetCpus = copts.cpusetCpus
	config.CpusetMems = copts.cpusetMems
	config.MemReservation = types.UnitBytes(copts.memoryReservation)
	config.MemSwapLimit = types.UnitBytes(copts.memorySwap)
	config.MemSwappiness = types.UnitBytes(copts.swappiness)
	config.ShmSize = types.UnitBytes(copts.shmSize)
	config.OomKillDisable = copts.oomKillDisable
	config.OomScoreAdj = int64(copts.oomScoreAdj)
	config.PidsLimit = copts.pidsLimit

	config.BlkioConfig = parseBlkioOpts(copts)
	ulimits, err := validateUlimitsOpts(copts.ulimits)
	if err != nil {
		return nil, err
	}
	config.Ulimits = ulimits

	// Namespace & CGroup
	networkMode := container.NetworkMode(copts.netMode.NetworkMode())
	config.NetworkMode = string(networkMode)

	pidMode := container.PidMode(copts.pidMode)
	if !pidMode.Valid() {
		return nil, fmt.Errorf("--pid: invalid PID mode")
	}
	config.PidMode = string(pidMode)

	utsMode := container.UTSMode(copts.utsMode)
	if !utsMode.Valid() {
		return nil, fmt.Errorf("--uts: invalid UTS mode")
	}
	config.UTSMode = string(utsMode)

	usernsMode := container.UsernsMode(copts.usernsMode)
	if !usernsMode.Valid() {
		return nil, fmt.Errorf("--userns: invalid USER mode")
	}
	config.UserMode = string(usernsMode)

	cgroupnsMode := container.CgroupnsMode(copts.cgroupnsMode)
	if !cgroupnsMode.Valid() {
		return nil, fmt.Errorf("--cgroupns: invalid CGROUP mode")
	}
	config.CgroupMode = string(cgroupnsMode)
	config.CgroupParent = copts.cgroupParent

	// Security
	config.User = copts.user
	config.GroupAdd = copts.groupAdd.GetAll()
	config.CapAdd = copts.capAdd.GetAll()
	config.CapDrop = copts.capDrop.GetAll()
	config.Privileged = copts.privileged
	config.SecurityOpt = copts.securityOpt.GetAll()

	securityOpts, err := parseSecurityOpts(copts.securityOpt.GetAll())
	if err != nil {
		return nil, err
	}
	config.SecurityOpt = securityOpts // TODO: parseSystemPaths https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L542
	config.Sysctls = copts.sysctls.GetAll()

	return &config, nil
}

func durationPtr(value time.Duration) *types.Duration {
	result := types.Duration(value)
	return &result
}

func intPtr(value int) *int {
	return &value
}

func uint32Ptr(value uint32) *uint32 {
	return &value
}

func uint64Ptr(value uint64) *uint64 {
	return &value
}
