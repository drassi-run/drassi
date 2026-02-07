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
	CreateRun(stage Stage) *ActionRun
}

type ActionRun struct {
	Run       ActionTask
	Stage     Stage
	Executor  ActionExecutor
	Condition workflows.Conditional
}

func (r *ActionRun) StepId() string {
	return r.StepSpec().Id
}

func (r *ActionRun) ActionSpec() ActionSpec {
	return r.Executor.ActionSpec()
}

func (r *ActionRun) StepSpec() *StepSpec {
	return r.Executor.StepExecutor().StepSpec()
}

func (r *ActionRun) JobSpec() *JobSpec {
	return r.Executor.StepExecutor().JobExecutor().JobSpec()
}

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type ActionTask func(context.Context) error

type ActionRunDecorator interface {
	DecorateActionRun(*ActionRun) ActionTask
}
