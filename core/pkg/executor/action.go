/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"

	"drassi.run/core/pkg/model/workflows"
	"go.uber.org/dig"
)

type ActionSpec interface {
	CreateExecutor(ctx context.Context, scope *dig.Scope, exec StepExecutor) (ActionExecutor, error)
}

type ActionExecutor interface {
	ActionSpec() ActionSpec
	StepExecutor() StepExecutor
	CreateTask(stage Stage) *ActionTask
}

type ActionTask struct {
	Run       ActionRun
	Stage     Stage
	Executor  ActionExecutor
	Condition workflows.Conditional
}

func (t *ActionTask) StepId() string {
	return t.StepSpec().Id
}

func (t *ActionTask) ActionSpec() ActionSpec {
	return t.Executor.ActionSpec()
}

func (t *ActionTask) StepSpec() *StepSpec {
	return t.Executor.StepExecutor().StepSpec()
}

func (t *ActionTask) JobSpec() *JobSpec {
	return t.Executor.StepExecutor().JobExecutor().JobSpec()
}

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type ActionRun func(context.Context) error

type ActionRunDecorator interface {
	DecorateActionRun(*ActionTask) ActionRun
}
