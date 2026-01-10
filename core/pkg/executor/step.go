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

type StepSpec struct {
	Id               string
	Uid              string
	Name             workflows.Evaluable[string]
	Condition        workflows.Conditional
	ContinueOnError  workflows.Evaluable[bool]
	TimeoutInMinutes workflows.Evaluable[int64]
	Env              workflows.Evaluable[map[string]string]
	Inputs           workflows.Evaluable[map[string]string]
	Outputs          workflows.Evaluable[map[string]string]

	// specific fields for each step type
	Def StepDef
}

type StepDef interface {
	PrepareExecute(ctx context.Context, scope *dig.Scope) (StepRun, error)
}

type StepRun interface {
	Def() StepDef

	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
}
