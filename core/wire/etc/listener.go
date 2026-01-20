/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package etc

import (
	"context"

	"drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

type stackListener struct {
	stack *stack
}

func (l *stackListener) OnInitializeJob(exec executor.JobExecutor, _ *dig.Scope) executor.EventHandler {
	return l.stack.jobChange(exec)
}

func (l *stackListener) OnRunJob(exec executor.JobExecutor) executor.EventHandler {
	return l.stack.jobChange(exec)
}

func (l *stackListener) OnRunStage(executor.JobExecutor, executor.Stage) executor.EventHandler {
	return l.stack.contextChange()
}

func (l *stackListener) OnFinalizeJob(exec executor.JobExecutor) executor.EventHandler {
	return l.stack.jobChange(exec)
}

func (l *stackListener) OnInitializeStep(exec executor.StepExecutor, _ *dig.Scope) executor.EventHandler {
	return l.stack.stepChange(exec)
}

func (l *stackListener) OnRunStep(exec executor.StepExecutor, _ executor.Stage) executor.EventHandler {
	return l.stack.stepChange(exec)
}

func (l *stackListener) OnRunTask(executor.StepExecutor, *executor.ActionRun) executor.EventHandler {
	return l.stack.contextChange()
}

type stack struct {
	job   executor.JobExecutor
	steps []executor.StepExecutor
	ctx   context.Context
}

func (s *stack) Context() context.Context {
	if ctx := s.ctx; ctx != nil {
		return ctx
	}
	return context.Background()
}

func (s *stack) Job() executor.JobExecutor {
	return s.job
}

func (s *stack) Root() executor.StepExecutor {
	if len(s.steps) == 0 {
		return nil
	}
	return s.steps[0]
}

func (s *stack) Leaf() executor.StepExecutor {
	if len(s.steps) == 0 {
		return nil
	}
	return s.steps[len(s.steps)-1]
}

func (s *stack) Stack() []executor.StepExecutor {
	return s.steps
}

func (s *stack) contextChange() executor.EventHandler {
	var prevCtx context.Context

	return &executor.FuncEventHandler{
		BeginFunc: func(ctx context.Context) error {
			prevCtx, s.ctx = s.ctx, ctx
			return nil
		},
		EndFunc: func(err error) error {
			s.ctx = prevCtx
			return nil
		},
	}
}

func (s *stack) jobChange(exec executor.JobExecutor) executor.EventHandler {
	var prevCtx context.Context

	return &executor.FuncEventHandler{
		BeginFunc: func(ctx context.Context) error {
			prevCtx, s.ctx = s.ctx, ctx
			s.job = exec
			return nil
		},
		EndFunc: func(err error) error {
			s.job = nil
			s.ctx = prevCtx
			return nil
		},
	}
}

func (s *stack) stepChange(exec executor.StepExecutor) executor.EventHandler {
	var prevCtx context.Context

	return &executor.FuncEventHandler{
		BeginFunc: func(ctx context.Context) error {
			prevCtx, s.ctx = s.ctx, ctx
			s.steps = append(s.steps, exec)
			return nil
		},
		EndFunc: func(err error) error {
			if len(s.steps) > 0 {
				s.steps = s.steps[:len(s.steps)-1]
			}
			s.ctx = prevCtx
			return nil
		},
	}
}
