package wire_cmdhandler

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	utilreader "drassi.run/core/pkg/util/reader"
)

var (
	ErrInvalidFile   = errors.New("invalid file")
	ErrNoJobRunning  = errors.New("no job is running")
	ErrNoStepRunning = errors.New("no step is running")
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func FileAddPath(sup executor.Supervisor) *command.FileHandler {
	run := func(r io.Reader) error {
		ctx := sup.Context()
		job := sup.Job()
		if job == nil {
			return ErrNoJobRunning
		}

		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("%w: un-expected file %q", ErrInvalidFile, hdr.Name)
			}

			if paths, err := readLine(reader); err != nil {
				return err
			} else {
				return job.AddPath(paths)
			}
		})
	}
	return command.NewFileHandler("GITHUB_PATH", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
func FileSetEnv(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, executor.StepExecutor.SetEnv)
	return command.NewFileHandler("GITHUB_ENV", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func FileSaveState(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, executor.StepExecutor.SaveState)
	return command.NewFileHandler("GITHUB_STATE", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func FileSetOutput(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, executor.StepExecutor.SetOutput)
	return command.NewFileHandler("GITHUB_OUTPUT", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
func CreateStepSummary(sup executor.Supervisor) *command.FileHandler {
	run := func(r io.Reader) error {
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}
		return step.CreateStepSummary(r)
	}
	return command.NewFileHandler("GITHUB_STEP_SUMMARY", run)
}

func stepCommandRun[R any](
	sup executor.Supervisor,
	trans func(r io.Reader) (R, error),
	set func(executor.StepExecutor, R) error,
) func(r io.Reader) error {
	return func(r io.Reader) error {
		ctx := sup.Context()
		step := sup.CurrentStep()
		if step == nil {
			return ErrNoStepRunning
		}

		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("%w: un-expected file %q", ErrInvalidFile, hdr.Name)
			}

			if res, err := trans(reader); err != nil {
				return err
			} else {
				return set(step, res)
			}
		})
	}
}
