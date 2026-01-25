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
	CreateExecutor(ctx context.Context, scope *dig.Scope) (ActionExecutor, error)
}

type ActionExecutor interface {
	ActionSpec() ActionSpec
	CreateRun(stage Stage) *ActionRun
}

type ActionRun struct {
	Condition workflows.Conditional
	Run       Task
}

type Task func(context.Context, StepExecutor) error

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type ActionRunDecorator interface {
	DecorateActionRun(Stage, *ActionRun) *ActionRun
}
