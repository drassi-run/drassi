/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"

	"go.uber.org/dig"
)

type NoopJobListener struct{}

func (l NoopJobListener) OnInitializeJob(JobExecutor, *dig.Scope) EventHandler { return nil }
func (l NoopJobListener) OnRunJob(JobExecutor) EventHandler                    { return nil }
func (l NoopJobListener) OnRunStage(JobExecutor, Stage) EventHandler           { return nil }
func (l NoopJobListener) OnFinalizeJob(JobExecutor) EventHandler               { return nil }

type NoopStepListener struct{}

func (l NoopStepListener) OnInitializeStep(StepExecutor, *dig.Scope) EventHandler { return nil }
func (l NoopStepListener) OnRunStep(StepExecutor, Stage) EventHandler             { return nil }
func (l NoopStepListener) OnRunTask(StepExecutor, *ActionRun) EventHandler        { return nil }

type FuncEventHandler struct {
	BeginFunc func(ctx context.Context) error
	EndFunc   func(err error) error
}

func (h *FuncEventHandler) Begin(ctx context.Context) error {
	if h.BeginFunc != nil {
		return h.BeginFunc(ctx)
	}
	return nil
}

func (h *FuncEventHandler) End(err error) error {
	if h.EndFunc != nil {
		return h.EndFunc(err)
	}
	return nil
}
