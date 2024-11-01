package cli

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/container/types"
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
	fm.mapStdio(copts)
	if err := fm.mapContainerSpec(flags, copts); err != nil {
		return nil, nil, err
	}

	return fm.Spec, fm.Stdio, nil
}

type flagMapper struct {
	Spec  *types.ContainerSpec
	Stdio *types.Stdio
}

// mapStdio function convert flags and copts to Stdio
func (fm *flagMapper) mapStdio(copts *containerOptions) {
	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L339-L358
	fm.Stdio = &types.Stdio{
		Tty:         copts.tty,
		Interactive: copts.stdin, // --interactive
	}
	stdio := fm.Stdio
	if copts.attach.Get("stdin") || copts.stdin {
		stdio.Attach |= types.Stdin
	}
	if copts.attach.Get("stdout") {
		stdio.Attach |= types.Stdout
	}
	if copts.attach.Get("stderr") {
		stdio.Attach |= types.Stderr
	}
	// If -a is not set, attach to stdout and stderr
	if copts.attach.Len() == 0 {
		stdio.Attach |= types.Stdout | types.Stderr
	}
}

// mapContainerSpec function convert flags and copts to Spec
// See also: https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L338-L740
func (fm *flagMapper) mapContainerSpec(flags *pflag.FlagSet, copts *containerOptions) error {
	fm.Spec = &types.ContainerSpec{
		Name:        copts.name,
		PullPolicy:  copts.pull,
		WorkingDir:  copts.workingDir,
		Environment: ConvertKVStringsToMap(copts.env.GetAll()),
		Labels:      ConvertKVStringsToMap(copts.labels.GetAll()),
		Annotations: copts.annotations.GetAll(),
	}
	spec := fm.Spec

	// https://github.com/docker/cli/blob/v27.3.1/cli/command/container/opts.go#L419-L424
	if copts.entrypoint != "" {
		spec.Entrypoint = []string{copts.entrypoint}
	} else if flags.Changed("entrypoint") {
		// if `--entrypoint=` is parsed then Entrypoint is reset
		spec.Entrypoint = []string{""}
	}

	if err := fm.mapNetwork(copts); err != nil {
		return err
	}
	if err := fm.mapStorage(copts); err != nil {
		return err
	}
	// TODO https://github.com/docker/cli/blob/26.0/cli/command/container/opts.go#L468-L493
	spec.Devices = copts.devices.GetAll()
	spec.DeviceCgroupRules = copts.deviceCgroupRules.GetAll()
	if err := fm.mapRuntime(flags, copts); err != nil {
		return err
	}
	if err := fm.mapResource(copts); err != nil {
		return err
	}
	if err := fm.mapSecurity(copts); err != nil {
		return err
	}

	return nil
}

// ConvertMapToKVString converts {"key":"value"} to ["key=value"]
func ConvertMapToKVString(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}

	r := make([]string, 0, len(m))
	for k, v := range m {
		r = append(r, fmt.Sprintf("%s=%s", k, v))
	}
	return r
}

// ConvertKVStringsToMap converts ["key=value"] to {"key":"value"}
func ConvertKVStringsToMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		k, v, _ := strings.Cut(value, "=")
		result[k] = v
	}

	return result
}
