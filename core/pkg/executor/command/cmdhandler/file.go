/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"context"
	"fmt"
	"io"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/command/issue"
	"drassi.run/core/pkg/scribe"
)

// FileAddPath create [command.FileHandler] that handle "GITHUB_PATH" command
//
//   - https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func FileAddPath[R SupportAddPath]() *command.FileHandler[R] {
	run := func(ctx context.Context, res R, r io.Reader) error {
		if paths, err := readLine(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for _, path := range paths {
				s.Debugf("Add path: %q", path)
			}
			res.AddPath(paths)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_PATH", run)
}

// FileSetEnv create [command.FileHandler] that handle "GITHUB_ENV" command
//
//   - https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
func FileSetEnv[R SupportSetEnv](reporter issue.Reporter) *command.FileHandler[R] {
	run := func(ctx context.Context, res R, r io.Reader) error {
		env, err := parseEnvVars(r)
		if err != nil {
			return err
		}

		s := scribe.FromContext(ctx)
		for k, v := range env {
			if setEnvBlockList.Has(k) {
				iss := &issue.Issue{
					Type:    issue.TypeError,
					Message: fmt.Sprintf("Can't update %q environment variable using '$GITHUB_ENV' command.", k),
				}
				if err := reporter.AddIssue(ctx, iss); err != nil {
					return err
				}
			} else {
				s.Debugf("Set env: %s = %s", k, v)
			}
		}

		res.SetEnv(env)
		return nil
	}
	return command.NewFileHandler("GITHUB_ENV", run)
}

// FileSaveState create [command.FileHandler] that handle "GITHUB_STATE" command
//
//   - https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func FileSaveState[R SupportSaveState]() *command.FileHandler[R] {
	run := func(ctx context.Context, res R, r io.Reader) error {
		if state, err := parseEnvVars(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for k, v := range state {
				s.Debugf("Save intra-action state: %s = %s", k, v)
			}
			res.SaveState(state)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_STATE", run)
}

// FileSetOutput create [command.FileHandler] that handle "GITHUB_OUTPUT" command
//
//   - https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func FileSetOutput[R SupportSetOutput]() *command.FileHandler[R] {
	run := func(ctx context.Context, res R, r io.Reader) error {
		if output, err := parseEnvVars(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for k, v := range output {
				s.Debugf("Set output: %s = %s", k, v)
			}
			res.SetOutput(output)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_OUTPUT", run)
}

// CreateStepSummary create [command.FileHandler] that handle "GITHUB_STEP_SUMMARY" command
//
//   - https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
func CreateStepSummary[R any]() *command.FileHandler[R] {
	run := func(ctx context.Context, _ R, r io.Reader) error {
		scribe.Debugf(ctx, "Create step summary")
		// TODO
		return nil
	}
	return command.NewFileHandler("GITHUB_STEP_SUMMARY", run)
}
