package cli

import (
	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/opts"
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
	if err := fm.mapRuntime(flags, copts); err != nil {
		return err
	}

	// Security
	if err := fm.mapSecurity(copts); err != nil {
		return err
	}

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
