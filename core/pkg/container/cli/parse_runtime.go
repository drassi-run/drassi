package cli

import (
	"fmt"
	"time"

	"drassi.run/core/pkg/container/types"
	"github.com/docker/cli/opts"
	"github.com/spf13/pflag"
)

func (fm *flagMapper) mapRuntime(flags *pflag.FlagSet, copts *containerOptions) error {
	fm.Spec.Runtime = copts.runtime
	//fm.Spec.Platform = copts.platform
	fm.Spec.Isolation = copts.isolation
	fm.Spec.StopSignal = copts.stopSignal
	if flags.Changed("stop-timeout") {
		d := time.Duration(copts.stopTimeout) * time.Second
		fm.Spec.StopTimeout = &d
	}

	if err := fm.mapLogging(copts); err != nil {
		return err
	}
	if err := fm.mapHealth(copts); err != nil {
		return err
	}

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
