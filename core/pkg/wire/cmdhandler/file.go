package wire_cmdhandler

import (
	"context"
	"errors"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/logging"
)

var (
	ErrInvalidFile   = errors.New("invalid file")
	ErrNoJobRunning  = errors.New("no job is running")
	ErrNoStepRunning = errors.New("no step is running")
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func FileAddPath(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		job := sup.Job()
		if job == nil {
			return ErrNoJobRunning
		}

		if paths, err := readLine(r); err != nil {
			return err
		} else {
			for _, path := range paths {
				logging.Debugf(l, "Add path: %q", path)
			}
			return job.AddPath(paths)
		}
	}
	return command.NewFileHandler("GITHUB_PATH", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
func FileSetEnv(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		if env, err := parseEnvVars(r); err != nil {
			return err
		} else {
			for k, v := range env {
				logging.Debugf(l, "Set env: %s = %s", k, v)
			}
			return step.SetEnv(env)
		}
	}
	return command.NewFileHandler("GITHUB_ENV", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func FileSaveState(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		if state, err := parseEnvVars(r); err != nil {
			return err
		} else {
			for k, v := range state {
				logging.Debugf(l, "Save intra-action state: %s = %s", k, v)
			}
			return step.SaveState(state)
		}
	}
	return command.NewFileHandler("GITHUB_STATE", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func FileSetOutput(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		if output, err := parseEnvVars(r); err != nil {
			return err
		} else {
			for k, v := range output {
				logging.Debugf(l, "Set output: %s = %s", k, v)
			}
			return step.SetOutput(output)
		}
	}
	return command.NewFileHandler("GITHUB_OUTPUT", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
func CreateStepSummary(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		logging.Debugf(l, "Create step summary")
		return step.CreateStepSummary(r)
	}
	return command.NewFileHandler("GITHUB_STEP_SUMMARY", run)
}
