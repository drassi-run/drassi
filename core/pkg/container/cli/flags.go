/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/cli/opts"
	"github.com/moby/moby/api/types/container"
	"github.com/spf13/pflag"
)

// Pull constants
const (
	PullImageAlways  = "always"
	PullImageMissing = "missing" // Default (matches previous behavior)
	PullImageNever   = "never"
)

// containerOptions is a data object with all the options for creating a container
// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L48-L146
//
//goland:noinspection SpellCheckingInspection
type containerOptions struct {
	attach              opts.ListOpts
	volumes             opts.ListOpts
	tmpfs               opts.ListOpts
	mounts              opts.MountOpt
	blkioWeightDevice   opts.WeightdeviceOpt
	deviceReadBps       opts.ThrottledeviceOpt
	deviceWriteBps      opts.ThrottledeviceOpt
	links               opts.ListOpts
	aliases             opts.ListOpts
	linkLocalIPs        opts.ListOpts
	deviceReadIOps      opts.ThrottledeviceOpt
	deviceWriteIOps     opts.ThrottledeviceOpt
	env                 opts.ListOpts
	labels              opts.ListOpts
	deviceCgroupRules   opts.ListOpts
	devices             opts.ListOpts
	gpus                opts.GpuOpts
	ulimits             *opts.UlimitOpt
	sysctls             *opts.MapOpts
	publish             opts.ListOpts
	expose              opts.ListOpts
	dns                 opts.ListOpts
	dnsSearch           opts.ListOpts
	dnsOptions          opts.ListOpts
	extraHosts          opts.ListOpts
	volumesFrom         opts.ListOpts
	envFile             opts.ListOpts
	capAdd              opts.ListOpts
	capDrop             opts.ListOpts
	groupAdd            opts.ListOpts
	securityOpt         opts.ListOpts
	storageOpt          opts.ListOpts
	labelsFile          opts.ListOpts
	loggingOpts         opts.ListOpts
	privileged          bool
	pidMode             string
	utsMode             string
	usernsMode          string
	cgroupnsMode        string
	publishAll          bool
	stdin               bool
	tty                 bool
	oomKillDisable      bool
	oomScoreAdj         int
	containerIDFile     string
	entrypoint          string
	hostname            string
	domainname          string
	memory              opts.MemBytes
	memoryReservation   opts.MemBytes
	memorySwap          opts.MemSwapBytes
	kernelMemory        opts.MemBytes
	user                string
	workingDir          string
	cpuCount            int64
	cpuShares           int64
	cpuPercent          int64
	cpuPeriod           int64
	cpuRealtimePeriod   int64
	cpuRealtimeRuntime  int64
	cpuQuota            int64
	cpus                opts.NanoCPUs
	cpusetCpus          string
	cpusetMems          string
	blkioWeight         uint16
	ioMaxBandwidth      opts.MemBytes
	ioMaxIOps           uint64
	swappiness          int64
	netMode             opts.NetworkOpt
	macAddress          string
	ipv4Address         string
	ipv6Address         string
	ipcMode             string
	pidsLimit           int64
	restartPolicy       string
	readonlyRootfs      bool
	loggingDriver       string
	cgroupParent        string
	volumeDriver        string
	stopSignal          string
	stopTimeout         int
	isolation           string
	shmSize             opts.MemBytes
	noHealthcheck       bool
	healthCmd           string
	healthInterval      time.Duration
	healthTimeout       time.Duration
	healthStartPeriod   time.Duration
	healthStartInterval time.Duration
	healthRetries       int
	runtime             string
	autoRemove          bool
	init                bool
	annotations         *opts.MapOpts

	//// options set from args
	//// see: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/create.go#L54-L57
	//Image string
	//Args  []string

	// options from createOptions
	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/create.go#L36-L42
	name string
	pull string // always, missing, never
}

// addFlags adds all command line flags that will be used by parse to the FlagSet
// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L149-L325
//
//goland:noinspection GoUnhandledErrorResult
func addFlags(flags *pflag.FlagSet) *containerOptions {
	copts := &containerOptions{
		aliases:           opts.NewListOpts(nil),
		attach:            opts.NewListOpts(validateAttach),
		blkioWeightDevice: opts.NewWeightdeviceOpt(opts.ValidateWeightDevice),
		capAdd:            opts.NewListOpts(nil),
		capDrop:           opts.NewListOpts(nil),
		dns:               opts.NewListOpts(opts.ValidateIPAddress),
		dnsOptions:        opts.NewListOpts(nil),
		dnsSearch:         opts.NewListOpts(opts.ValidateDNSSearch),
		deviceCgroupRules: opts.NewListOpts(validateDeviceCgroupRule),
		deviceReadBps:     opts.NewThrottledeviceOpt(opts.ValidateThrottleBpsDevice),
		deviceReadIOps:    opts.NewThrottledeviceOpt(opts.ValidateThrottleIOpsDevice),
		deviceWriteBps:    opts.NewThrottledeviceOpt(opts.ValidateThrottleBpsDevice),
		deviceWriteIOps:   opts.NewThrottledeviceOpt(opts.ValidateThrottleIOpsDevice),
		devices:           opts.NewListOpts(nil), // devices can only be validated after we know the server OS
		env:               opts.NewListOpts(opts.ValidateEnv),
		envFile:           opts.NewListOpts(nil),
		expose:            opts.NewListOpts(nil),
		extraHosts:        opts.NewListOpts(opts.ValidateExtraHost),
		groupAdd:          opts.NewListOpts(nil),
		labels:            opts.NewListOpts(opts.ValidateLabel),
		labelsFile:        opts.NewListOpts(nil),
		linkLocalIPs:      opts.NewListOpts(nil),
		links:             opts.NewListOpts(opts.ValidateLink),
		loggingOpts:       opts.NewListOpts(nil),
		publish:           opts.NewListOpts(nil),
		securityOpt:       opts.NewListOpts(nil),
		storageOpt:        opts.NewListOpts(nil),
		sysctls:           opts.NewMapOpts(nil, opts.ValidateSysctl),
		tmpfs:             opts.NewListOpts(nil),
		ulimits:           opts.NewUlimitOpt(nil),
		volumes:           opts.NewListOpts(nil),
		volumesFrom:       opts.NewListOpts(nil),
		annotations:       opts.NewMapOpts(nil, nil),
	}

	// General purpose flags
	flags.VarP(&copts.attach, "attach", "a", "Attach to STDIN, STDOUT or STDERR")
	flags.Var(&copts.deviceCgroupRules, "device-cgroup-rule", "Add a rule to the cgroup allowed devices list")
	flags.Var(&copts.devices, "device", "Add a host device to the container")
	flags.Var(&copts.gpus, "gpus", "GPU devices to add to the container ('all' to pass all GPUs)")
	flags.SetAnnotation("gpus", "version", []string{"1.40"})
	flags.VarP(&copts.env, "env", "e", "Set environment variables")
	flags.Var(&copts.envFile, "env-file", "Read in a file of environment variables")
	flags.StringVar(&copts.entrypoint, "entrypoint", "", "Overwrite the default ENTRYPOINT of the image")
	flags.Var(&copts.groupAdd, "group-add", "Add additional groups to join")
	flags.StringVarP(&copts.hostname, "hostname", "h", "", "Container host name")
	flags.StringVar(&copts.domainname, "domainname", "", "Container NIS domain name")
	flags.BoolVarP(&copts.stdin, "interactive", "i", false, "Keep STDIN open even if not attached")
	flags.VarP(&copts.labels, "label", "l", "Set meta data on a container")
	flags.Var(&copts.labelsFile, "label-file", "Read in a line delimited file of labels")
	flags.BoolVar(&copts.readonlyRootfs, "read-only", false, "Mount the container's root filesystem as read only")
	flags.StringVar(&copts.restartPolicy, "restart", string(container.RestartPolicyDisabled), "Restart policy to apply when a container exits")
	flags.StringVar(&copts.stopSignal, "stop-signal", "", "Signal to stop the container")
	flags.IntVar(&copts.stopTimeout, "stop-timeout", 0, "Timeout (in seconds) to stop a container")
	flags.SetAnnotation("stop-timeout", "version", []string{"1.25"})
	flags.Var(copts.sysctls, "sysctl", "Sysctl options")
	flags.BoolVarP(&copts.tty, "tty", "t", false, "Allocate a pseudo-TTY")
	flags.Var(copts.ulimits, "ulimit", "Ulimit options")
	flags.StringVarP(&copts.user, "user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
	flags.StringVarP(&copts.workingDir, "workdir", "w", "", "Working directory inside the container")
	flags.BoolVar(&copts.autoRemove, "rm", false, "Automatically remove the container and its associated anonymous volumes when it exits")

	// Security
	flags.Var(&copts.capAdd, "cap-add", "Add Linux capabilities")
	flags.Var(&copts.capDrop, "cap-drop", "Drop Linux capabilities")
	flags.BoolVar(&copts.privileged, "privileged", false, "Give extended privileges to this container")
	flags.Var(&copts.securityOpt, "security-opt", "Security Options")
	flags.StringVar(&copts.usernsMode, "userns", "", "User namespace to use")
	flags.StringVar(&copts.cgroupnsMode, "cgroupns", "", `Cgroup namespace to use (host|private)
'host':    Run the container in the Docker host's cgroup namespace
'private': Run the container in its own private cgroup namespace
'':        Use the cgroup namespace as configured by the
           default-cgroupns-mode option on the daemon (default)`)
	flags.SetAnnotation("cgroupns", "version", []string{"1.41"})

	// Network and port publishing flag
	flags.Var(&copts.extraHosts, "add-host", "Add a custom host-to-IP mapping (host:ip)")
	flags.Var(&copts.dns, "dns", "Set custom DNS servers")
	// We allow for both "--dns-opt" and "--dns-option", although the latter is the recommended way.
	// This is to be consistent with service create/update
	flags.Var(&copts.dnsOptions, "dns-opt", "Set DNS options")
	flags.Var(&copts.dnsOptions, "dns-option", "Set DNS options")
	flags.MarkHidden("dns-opt")
	flags.Var(&copts.dnsSearch, "dns-search", "Set custom DNS search domains")
	flags.Var(&copts.expose, "expose", "Expose a port or a range of ports")
	flags.StringVar(&copts.ipv4Address, "ip", "", "IPv4 address (e.g., 172.30.100.104)")
	flags.StringVar(&copts.ipv6Address, "ip6", "", "IPv6 address (e.g., 2001:db8::33)")
	flags.Var(&copts.links, "link", "Add link to another container")
	flags.Var(&copts.linkLocalIPs, "link-local-ip", "Container IPv4/IPv6 link-local addresses")
	flags.StringVar(&copts.macAddress, "mac-address", "", "Container MAC address (e.g., 92:d0:c6:0a:29:33)")
	flags.VarP(&copts.publish, "publish", "p", "Publish a container's port(s) to the host")
	flags.BoolVarP(&copts.publishAll, "publish-all", "P", false, "Publish all exposed ports to random ports")
	// We allow for both "--net" and "--network", although the latter is the recommended way.
	flags.Var(&copts.netMode, "net", "Connect a container to a network")
	flags.Var(&copts.netMode, "network", "Connect a container to a network")
	flags.MarkHidden("net")
	// We allow for both "--net-alias" and "--network-alias", although the latter is the recommended way.
	flags.Var(&copts.aliases, "net-alias", "Add network-scoped alias for the container")
	flags.Var(&copts.aliases, "network-alias", "Add network-scoped alias for the container")
	flags.MarkHidden("net-alias")

	// Logging and storage
	flags.StringVar(&copts.loggingDriver, "log-driver", "", "Logging driver for the container")
	flags.StringVar(&copts.volumeDriver, "volume-driver", "", "Optional volume driver for the container")
	flags.Var(&copts.loggingOpts, "log-opt", "Log driver options")
	flags.Var(&copts.storageOpt, "storage-opt", "Storage driver options for the container")
	flags.Var(&copts.tmpfs, "tmpfs", "Mount a tmpfs directory")
	flags.Var(&copts.volumesFrom, "volumes-from", "Mount volumes from the specified container(s)")
	flags.VarP(&copts.volumes, "volume", "v", "Bind mount a volume")
	flags.Var(&copts.mounts, "mount", "Attach a filesystem mount to the container")

	// Health-checking
	flags.StringVar(&copts.healthCmd, "health-cmd", "", "Command to run to check health")
	flags.DurationVar(&copts.healthInterval, "health-interval", 0, "Time between running the check (ms|s|m|h) (default 0s)")
	flags.IntVar(&copts.healthRetries, "health-retries", 0, "Consecutive failures needed to report unhealthy")
	flags.DurationVar(&copts.healthTimeout, "health-timeout", 0, "Maximum time to allow one check to run (ms|s|m|h) (default 0s)")
	flags.DurationVar(&copts.healthStartPeriod, "health-start-period", 0, "Start period for the container to initialize before starting health-retries countdown (ms|s|m|h) (default 0s)")
	flags.SetAnnotation("health-start-period", "version", []string{"1.29"})
	flags.DurationVar(&copts.healthStartInterval, "health-start-interval", 0, "Time between running the check during the start period (ms|s|m|h) (default 0s)")
	flags.SetAnnotation("health-start-interval", "version", []string{"1.44"})
	flags.BoolVar(&copts.noHealthcheck, "no-healthcheck", false, "Disable any container-specified HEALTHCHECK")

	// Resource management
	flags.Uint16Var(&copts.blkioWeight, "blkio-weight", 0, "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)")
	flags.Var(&copts.blkioWeightDevice, "blkio-weight-device", "Block IO weight (relative device weight)")
	flags.StringVar(&copts.containerIDFile, "cidfile", "", "Write the container ID to the file")
	flags.StringVar(&copts.cpusetCpus, "cpuset-cpus", "", "CPUs in which to allow execution (0-3, 0,1)")
	flags.StringVar(&copts.cpusetMems, "cpuset-mems", "", "MEMs in which to allow execution (0-3, 0,1)")
	flags.Int64Var(&copts.cpuCount, "cpu-count", 0, "CPU count (Windows only)")
	flags.SetAnnotation("cpu-count", "ostype", []string{"windows"})
	flags.Int64Var(&copts.cpuPercent, "cpu-percent", 0, "CPU percent (Windows only)")
	flags.SetAnnotation("cpu-percent", "ostype", []string{"windows"})
	flags.Int64Var(&copts.cpuPeriod, "cpu-period", 0, "Limit CPU CFS (Completely Fair Scheduler) period")
	flags.Int64Var(&copts.cpuQuota, "cpu-quota", 0, "Limit CPU CFS (Completely Fair Scheduler) quota")
	flags.Int64Var(&copts.cpuRealtimePeriod, "cpu-rt-period", 0, "Limit CPU real-time period in microseconds")
	flags.SetAnnotation("cpu-rt-period", "version", []string{"1.25"})
	flags.Int64Var(&copts.cpuRealtimeRuntime, "cpu-rt-runtime", 0, "Limit CPU real-time runtime in microseconds")
	flags.SetAnnotation("cpu-rt-runtime", "version", []string{"1.25"})
	flags.Int64VarP(&copts.cpuShares, "cpu-shares", "c", 0, "CPU shares (relative weight)")
	flags.Var(&copts.cpus, "cpus", "Number of CPUs")
	flags.SetAnnotation("cpus", "version", []string{"1.25"})
	flags.Var(&copts.deviceReadBps, "device-read-bps", "Limit read rate (bytes per second) from a device")
	flags.Var(&copts.deviceReadIOps, "device-read-iops", "Limit read rate (IO per second) from a device")
	flags.Var(&copts.deviceWriteBps, "device-write-bps", "Limit write rate (bytes per second) to a device")
	flags.Var(&copts.deviceWriteIOps, "device-write-iops", "Limit write rate (IO per second) to a device")
	flags.Var(&copts.ioMaxBandwidth, "io-maxbandwidth", "Maximum IO bandwidth limit for the system drive (Windows only)")
	flags.SetAnnotation("io-maxbandwidth", "ostype", []string{"windows"})
	flags.Uint64Var(&copts.ioMaxIOps, "io-maxiops", 0, "Maximum IOps limit for the system drive (Windows only)")
	flags.SetAnnotation("io-maxiops", "ostype", []string{"windows"})
	flags.Var(&copts.kernelMemory, "kernel-memory", "Kernel memory limit")
	flags.VarP(&copts.memory, "memory", "m", "Memory limit")
	flags.Var(&copts.memoryReservation, "memory-reservation", "Memory soft limit")
	flags.Var(&copts.memorySwap, "memory-swap", "Swap limit equal to memory plus swap: '-1' to enable unlimited swap")
	flags.Int64Var(&copts.swappiness, "memory-swappiness", -1, "Tune container memory swappiness (0 to 100)")
	flags.BoolVar(&copts.oomKillDisable, "oom-kill-disable", false, "Disable OOM Killer")
	flags.IntVar(&copts.oomScoreAdj, "oom-score-adj", 0, "Tune host's OOM preferences (-1000 to 1000)")
	flags.Int64Var(&copts.pidsLimit, "pids-limit", 0, "Tune container pids limit (set -1 for unlimited)")

	// Low-level execution (cgroups, namespaces, ...)
	flags.StringVar(&copts.cgroupParent, "cgroup-parent", "", "Optional parent cgroup for the container")
	flags.StringVar(&copts.ipcMode, "ipc", "", "IPC mode to use")
	flags.StringVar(&copts.isolation, "isolation", "", "Container isolation technology")
	flags.StringVar(&copts.pidMode, "pid", "", "PID namespace to use")
	flags.Var(&copts.shmSize, "shm-size", "Size of /dev/shm")
	flags.StringVar(&copts.utsMode, "uts", "", "UTS namespace to use")
	flags.StringVar(&copts.runtime, "runtime", "", "Runtime to use for this container")

	flags.BoolVar(&copts.init, "init", false, "Run an init inside the container that forwards signals and reaps processes")
	flags.SetAnnotation("init", "version", []string{"1.25"})

	flags.Var(copts.annotations, "annotation", "Add an annotation to the container (passed through to the OCI runtime)")
	flags.SetAnnotation("annotation", "version", []string{"1.43"})

	// options from createOptions
	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/create.go#L69-L70
	flags.StringVar(&copts.name, "name", "", "Assign a name to the container")
	flags.StringVar(&copts.pull, "pull", PullImageMissing, `Pull image before creating ("`+PullImageAlways+`", "|`+PullImageMissing+`", "`+PullImageNever+`")`)

	return copts
}

var deviceCgroupRuleRegexp = regexp.MustCompile(`^[acb] ([0-9]+|\*):([0-9]+|\*) [rwm]{1,3}$`)

// validateDeviceCgroupRule validates a device cgroup rule string format
// It will make sure 'val' is in the form:
//
//	'type major:minor mode'
func validateDeviceCgroupRule(val string) (string, error) {
	if deviceCgroupRuleRegexp.MatchString(val) {
		return val, nil
	}

	return val, fmt.Errorf("invalid device cgroup format %q", val)
}

// validateAttach validates that the specified string is a valid attach option.
func validateAttach(val string) (string, error) {
	s := strings.ToLower(val)
	for _, str := range []string{"stdin", "stdout", "stderr"} {
		if s == str {
			return s, nil
		}
	}
	return val, fmt.Errorf("valid streams are STDIN, STDOUT and STDERR")
}
