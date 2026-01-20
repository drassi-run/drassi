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

type Listener interface {
	JobListener
	StepListener
}

type JobListener interface {
	OnInitializeJob(exec JobExecutor, scope *dig.Scope) EventHandler
	OnRunJob(exec JobExecutor) EventHandler
	OnRunStage(exec JobExecutor, stage Stage) EventHandler
	OnFinalizeJob(exec JobExecutor) EventHandler
}

type StepListener interface {
	OnInitializeStep(exec StepExecutor, scope *dig.Scope) EventHandler
	OnRunStep(exec StepExecutor, stage Stage) EventHandler
	OnRunTask(exec StepExecutor, task *ActionRun) EventHandler
}

type EventHandler interface {
	Begin(ctx context.Context) error
	End(err error) error
}
