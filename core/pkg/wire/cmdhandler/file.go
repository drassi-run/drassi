/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_cmdhandler

import (
	"context"
	"errors"
	"fmt"
	"io"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/scribe"
)

var (
	ErrInvalidFile   = errors.New("invalid file")
	ErrNoJobRunning  = errors.New("no job is running")
	ErrNoStepRunning = errors.New("no step is running")
)

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func FileAddPath(stack executor.Stack) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		job := stack.Job()
		if job == nil {
			return ErrNoJobRunning
		}

		if paths, err := readLine(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for _, path := range paths {
				s.Debugf("Add path: %q", path)
			}
			job.AddPath(paths)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_PATH", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
func FileSetEnv(stack executor.Stack, tracker support.Tracker) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		env, err := parseEnvVars(r)
		if err != nil {
			return err
		}

		s := scribe.FromContext(ctx)
		for k, v := range env {
			if setEnvBlockList.Has(k) {
				iss := &support.Issue{
					Type:    support.IssueTypeError,
					Message: fmt.Sprintf("Can't update %q environment variable using '$GITHUB_ENV' command.", k),
				}
				if err := tracker.AddIssue(ctx, iss); err != nil {
					return err
				}
			} else {
				s.Debugf("Set env: %s = %s", k, v)
			}
		}

		for _, step := range stack.Stack() {
			step.SetEnv(env)
		}
		stack.Job().SetEnv(env)
		return nil
	}
	return command.NewFileHandler("GITHUB_ENV", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
func FileSaveState(stack executor.Stack) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := stack.Root()
		if step == nil {
			return ErrNoStepRunning
		}

		if state, err := parseEnvVars(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for k, v := range state {
				s.Debugf("Save intra-action state: %s = %s", k, v)
			}
			step.SaveState(state)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_STATE", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
func FileSetOutput(stack executor.Stack) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := stack.Leaf()
		if step == nil {
			return ErrNoStepRunning
		}

		if output, err := parseEnvVars(r); err != nil {
			return err
		} else {
			s := scribe.FromContext(ctx)
			for k, v := range output {
				s.Debugf("Set output: %s = %s", k, v)
			}
			step.SetOutput(output)
			return nil
		}
	}
	return command.NewFileHandler("GITHUB_OUTPUT", run)
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
func CreateStepSummary(stack executor.Stack) *command.FileHandler {
	run := func(ctx context.Context, r io.Reader) error {
		step := stack.Leaf()
		if step == nil {
			return ErrNoStepRunning
		}

		scribe.Debugf(ctx, "Create step summary")
		return step.CreateStepSummary(r)
	}
	return command.NewFileHandler("GITHUB_STEP_SUMMARY", run)
}
