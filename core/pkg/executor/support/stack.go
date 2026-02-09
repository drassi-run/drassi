/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package support

import (
	"context"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/model/records"
	xtypes "drassi.run/core/util/types"
)

type Stack struct {
	job      *xtypes.Pair[executor.Stage, executor.JobExecutor]
	currStep *xtypes.Pair[executor.Stage, executor.StepExecutor]
}

func NewStack() *Stack {
	return new(Stack)
}

func (s *Stack) Job() (executor.Stage, executor.JobExecutor) {
	if job := s.job; job != nil {
		return job.Key, job.Value
	}
	return "", nil
}

func (s *Stack) CurrentStep() (executor.Stage, executor.StepExecutor) {
	if step := s.currStep; step != nil {
		return step.Key, step.Value
	}
	return "", nil
}

func (s *Stack) DecorateJobRun(job *executor.JobRun) executor.JobTask {
	run := job.Run
	return func(ctx context.Context) (*records.Job, error) {
		prev := s.job
		s.job = xtypes.NewPair(job.Stage, job.Executor)
		defer func() { s.job = prev }()

		return run(ctx)
	}
}

func (s *Stack) DecorateStepRun(step *executor.StepRun) executor.StepTask {
	run := step.Run
	return func(ctx context.Context) (*records.Step, error) {
		prev := s.currStep
		s.currStep = xtypes.NewPair(step.Stage, step.Executor)
		defer func() { s.currStep = prev }()

		return run(ctx)
	}
}
