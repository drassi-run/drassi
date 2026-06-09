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
	Action ActionSpec
}

func (spec *StepSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope,
	exec JobExecutor, parent StepExecutor,
) (StepExecutor, error) {
	e := &stepExecutor{spec: spec, jExec: exec, parent: parent}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}
