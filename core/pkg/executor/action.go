/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"errors"

	"drassi.run/core/pkg/model/records"
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

	Name() workflows.Evaluable[string]
	Env() workflows.Evaluable[map[string]string]
	Inputs() workflows.Evaluable[map[string]string]
	Outputs() workflows.Evaluable[map[string]string]
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

type ActionRun func(context.Context) (records.Result, error)

func runActionE(fn func(context.Context) error) ActionRun {
	return func(ctx context.Context) (records.Result, error) {
		err := fn(ctx)
		if err == nil {
			return records.ResultSuccess, nil
		}
		if errors.Is(err, context.Canceled) {
			cause := context.Cause(ctx)
			return records.ResultCancelled, cause
		}
		return records.ResultFailure, err
	}
}
