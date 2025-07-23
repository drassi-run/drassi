/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"fmt"
	"slices"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/util/dig"
	"go.uber.org/dig"
)

type CompositeStepRun struct {
	BaseStepRun
	StepRuns []StepRun

	children map[string]StepExecutor
	exprEnv  expression.Env
	inputs   map[string]string
}

func (sr *CompositeStepRun) Initialize(ctx context.Context, scope *dig.Scope) error {
	var exec StepExecutor
	if err := xdig.Populate(scope, &exec); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.exprEnv); err != nil {
		return err
	}

	// create a new intermediate scope to store composite values (inputs & exprEnv)
	sr.inputs = make(map[string]string)
	scope = scope.Scope("composite")
	opts := []expression.Option{
		// inputs from upper layers will NOT be passed to child steps
		expression.WithVariable("inputs", sr.inputs),
		expression.WithLibrary(libraries.StatusLib(exec)),
	}
	if exprEnv, err := sr.exprEnv.New(opts...); err != nil {
		return err
	} else if err = xdig.Supply(scope, exprEnv); err != nil {
		return err
	}
	if err := xdig.Supply(scope, sr.inputs); err != nil {
		return err
	}

	// TODO: concurrent version of Initialize is temporary disable because of concurrent map writes in scope
	//g, ctx := errgroup.WithContext(exec.Context())
	//for _, step := range sr.StepRuns {
	//	cExec := exec.NewChildExecutor(step)
	//	cScope := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
	//	g.Go(func() error {
	//		cExec.SetContext(ctx)
	//		return cExec.Initialize(cScope)
	//	})
	//}
	//return g.Wait()

	sr.children = make(map[string]StepExecutor, len(sr.StepRuns))
	for _, step := range sr.StepRuns {
		cExec := NewStepExecutor(step)
		sr.children[step.StepId()] = cExec

		cScope := scope.Scope(fmt.Sprintf("step(%s)", step.StepId()))
		if err := cExec.Initialize(ctx, cScope); err != nil {
			return err
		}
	}
	return nil
}

func (sr *CompositeStepRun) PreTask() *Task {
	return sr.createStageTask(StagePre)
}

func (sr *CompositeStepRun) MainTask() *Task {
	return sr.createStageTask(StageMain)
}

func (sr *CompositeStepRun) PostTask() *Task {
	return sr.createStageTask(StagePost)
}

func (sr *CompositeStepRun) createStageTask(stage Stage) *Task {
	taskIds := make([]string, len(sr.StepRuns))
	for i, step := range sr.StepRuns {
		taskIds[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(taskIds) // in-place reverse
	}

	taskRun := func(ctx context.Context, exec StepExecutor) error {
		if err := sr.computeInputs(); err != nil {
			return err
		}

		for _, id := range taskIds {
			cExec := sr.children[id]
			if cExec == nil {
				return fmt.Errorf("task %q has no child context", id)
			}

			res := cExec.RunStep(ctx, stage)
			if res != nil && res.Conclusion == records.ResultFailure {
				exec.SetStatus(records.ResultFailure)
				return fmt.Errorf("step %q (%s) failed", id, stage)
			}
		}

		return sr.produceOutputs(exec)
	}

	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L472-L490
	var condition workflows.Conditional
	if stage == StagePre || stage == StagePost {
		condition = "always()"
	}

	return &Task{
		StepId:    sr.Id,
		Stage:     stage,
		Condition: condition,
		Run:       taskRun,
	}
}

func (sr *CompositeStepRun) computeInputs() error {
	clear(sr.inputs)
	return evaluator.Evaluate(sr.exprEnv, sr.Inputs, &sr.inputs)
}

func (sr *CompositeStepRun) produceOutputs(exec StepExecutor) error {
	outputs := make(map[string]string)
	if err := evaluator.Evaluate(sr.exprEnv, sr.Outputs, &outputs); err != nil {
		return err
	}
	exec.SetOutput(outputs)
	return nil
}
