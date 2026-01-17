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

type JobSpec struct {
	Id   string
	Uid  string
	Name workflows.Evaluable[string]

	Container workflows.Evaluable[*workflows.Container]
	Services  workflows.Evaluable[map[string]*workflows.Container]

	Env      workflows.Evaluable[map[string]string]
	Inputs   workflows.Evaluable[map[string]string]
	Outputs  workflows.Evaluable[map[string]string]
	Defaults workflows.Evaluable[workflows.Defaults]
	// Environment

	Steps []*StepSpec
}

func (spec *JobSpec) CreateExecutor(ctx context.Context, scope *dig.Scope) (JobExecutor, error) {
	e := &jobExecutor{spec: spec}
	return e, nil
}
