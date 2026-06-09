/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

type JobRunDecorator interface {
	DecorateJobRun(*JobTask) JobRun
}

type StepRunDecorator interface {
	DecorateStepRun(*StepTask) StepRun
}

type ActionRunDecorator interface {
	DecorateActionRun(*ActionTask) ActionRun
}

type MultiJobRunDecorator []JobRunDecorator

func (dec MultiJobRunDecorator) DecorateJobRun(task *JobTask) JobRun {
	for _, d := range dec {
		task.Run = d.DecorateJobRun(task)
	}
	return task.Run
}

type MultiStepRunDecorator []StepRunDecorator

func (dec MultiStepRunDecorator) DecorateStepRun(task *StepTask) StepRun {
	for _, d := range dec {
		task.Run = d.DecorateStepRun(task)
	}
	return task.Run
}

type MultiActionRunDecorator []ActionRunDecorator

func (dec MultiActionRunDecorator) DecorateActionRun(task *ActionTask) ActionRun {
	for _, d := range dec {
		task.Run = d.DecorateActionRun(task)
	}
	return task.Run
}
