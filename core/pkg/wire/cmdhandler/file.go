package wire_cmdhandler

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/util/tar"
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

		found := false
		return xtar.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if !xtar.IsRegular(hdr) {
				return fmt.Errorf("%w: un-expected %s file", ErrInvalidFile, xtar.FileType(hdr.Typeflag))
			}
			if found {
				return fmt.Errorf("%w: un-expected multiple files", ErrInvalidFile)
			}
			found = true

			if paths, err := readLine(reader); err != nil {
				return err
			} else {
				for _, path := range paths {
					logging.Debugf(l, "Add path: %q", path)
				}
				return job.AddPath(paths)
			}
		})
	}
	return command.NewFileHandler("GITHUB_PATH", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
func FileSetEnv(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, func(exec executor.StepExecutor, env map[string]string) error {
		for k, v := range env {
			logging.Debugf(l, "Set env: %s = %s", k, v)
		}
		return exec.SetEnv(env)
	})
	return command.NewFileHandler("GITHUB_ENV", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func FileSaveState(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, func(exec executor.StepExecutor, state map[string]string) error {
		for k, v := range state {
			logging.Debugf(l, "Save intra-action state: %s = %s", k, v)
		}
		return exec.SaveState(state)
	})
	return command.NewFileHandler("GITHUB_STATE", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func FileSetOutput(sup executor.Supervisor, l logging.Logger) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, func(exec executor.StepExecutor, output map[string]string) error {
		for k, v := range output {
			logging.Debugf(l, "Save intra-action state: %s = %s", k, v)
		}
		return exec.SetOutput(output)
	})
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

func stepCommandRun[R any](
	sup executor.Supervisor,
	trans func(r io.Reader) (R, error),
	set func(executor.StepExecutor, R) error,
) func(ctx context.Context, r io.Reader) error {
	return func(ctx context.Context, r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		found := false
		return xtar.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if !xtar.IsRegular(hdr) {
				return fmt.Errorf("%w: un-expected %s file", ErrInvalidFile, xtar.FileType(hdr.Typeflag))
			}
			if found {
				return fmt.Errorf("%w: un-expected multiple files", ErrInvalidFile)
			}
			found = true

			if res, err := trans(reader); err != nil {
				return err
			} else {
				return set(step, res)
			}
		})
	}
}
