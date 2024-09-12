package cmdhandler

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	utilreader "drassi.run/core/pkg/util/reader"
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func fileAddPath(sup executor.Supervisor) *command.FileHandler {
	run := func(r io.Reader) error {
		ctx := sup.Context()
		job := sup.Job()
		if job == nil {
			return errors.New("no job found")
		}

		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
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
func fileSetEnv(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, func(step executor.StepExecutor, m map[string]string) error {
		return step.SetEnv(m, true)
	})
	return command.NewFileHandler("GITHUB_ENV", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func fileSaveState(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, executor.StepExecutor.SaveState)
	return command.NewFileHandler("GITHUB_STATE", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func fileSetOutput(sup executor.Supervisor) *command.FileHandler {
	run := stepCommandRun(sup, parseEnvVars, executor.StepExecutor.SetOutput)
	return command.NewFileHandler("GITHUB_OUTPUT", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
func createStepSummary(sup executor.Supervisor) *command.FileHandler {
	run := func(r io.Reader) error {
		// TODO: implement GITHUB_STEP_SUMMARY file command
		return nil
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
			return errors.New("no step found")
		}

		return utilreader.Untar(ctx, r, func(hdr *tar.Header, reader io.Reader) error {
			if hdr.Name != "" {
				return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
			}

			if res, err := trans(reader); err != nil {
				return err
			} else {
				return set(step, res)
			}
		})
	}
}
