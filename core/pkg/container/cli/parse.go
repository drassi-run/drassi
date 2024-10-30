package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"drassi.run/core/pkg/container/types"
	composetypes "github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/cli/opts"
	docker "github.com/docker/docker/api/types/container"
	"github.com/google/shlex"
	"github.com/spf13/pflag"
)

func Parse(opts string) (*types.ContainerSpec, *types.Stdio, error) {
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
	Spec  types.ContainerSpec
	Stdio types.Stdio
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
	if err := fm.mapNetwork(copts); err != nil {
		return err
	}

	// Storage & Device
	if err := fm.mapStorage(copts); err != nil {
		return err
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
		fm.Stdio.Attach |= types.Stdin
	}
	if copts.attach.Get("stdout") {
		fm.Stdio.Attach |= types.Stdout
	}
	if copts.attach.Get("stderr") {
		fm.Stdio.Attach |= types.Stderr
	}
	// If -a is not set, attach to stdout and stderr
	if copts.attach.Len() == 0 {
		fm.Stdio.Attach |= types.Stdout | types.Stderr
	}
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

	fm.Spec.HealthCheck = &types.HealthCheckConfig{
		Test:          probe,
		Timeout:       copts.healthTimeout,
		Interval:      copts.healthInterval,
		Retries:       copts.healthRetries,
		StartPeriod:   copts.healthStartPeriod,
		StartInterval: copts.healthStartInterval,
	}

	return nil
}

func durationPtr(value time.Duration) *composetypes.Duration {
	result := composetypes.Duration(value)
	return &result
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
