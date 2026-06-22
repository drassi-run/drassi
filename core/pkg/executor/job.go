/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"drassi.run/core/pkg/model/workflows"
	"go.uber.org/dig"
)

type JobSpec struct {
	Id   string                      // job unique identifier from workflow file
	Uid  string                      // job UUID, auto generated
	Ref  string                      // job reference, combined from strategy.matrix or '__default'
	Name workflows.Evaluable[string] // name of the job, which is displayed in the Web UI

	Container workflows.Evaluable[*workflows.Container]
	Services  workflows.Evaluable[map[string]*workflows.Container]

	Env      workflows.Evaluable[map[string]string]
	Inputs   workflows.Evaluable[map[string]string]
	Outputs  workflows.Evaluable[map[string]string]
	Defaults workflows.Evaluable[workflows.Defaults]
	// Environment

	Steps []*StepSpec
}

func (spec *JobSpec) CreateExecutor(scope *dig.Scope) JobExecutor {
	return &jobExecutor{spec: spec, scope: scope}
}
