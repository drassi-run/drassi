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
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

type CompositeActionSpec struct {
	Inputs  workflows.Evaluable[map[string]string]
	Outputs workflows.Evaluable[map[string]string]

	Steps []*StepSpec
}

func (spec *CompositeActionSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope, exec StepExecutor,
) (ActionExecutor, error) {
	e := &compositeActionExecutor{spec: spec, sExec: exec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type compositeActionExecutor struct {
	spec  *CompositeActionSpec
	sExec StepExecutor

	decorator StepRunDecorator
	exprEnv   expression.Env
	inputs    map[string]string
	children  map[string]StepExecutor
}

func (e *compositeActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.decorator); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}

	// create a new intermediate scope to store composite values (inputs & exprEnv)
	e.inputs = make(map[string]string)
	scope = scope.Scope("composite")
	opts := []expression.Option{
		// inputs from upper layers will NOT be passed to child steps
		expression.WithVariable("inputs", e.inputs),
		expression.WithLibrary(libraries.StatusLib(e.sExec)),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
		return err
	} else if err = xdig.Supply(scope, exprEnv); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.inputs); err != nil {
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

	e.children = make(map[string]StepExecutor, len(e.spec.Steps))
	for _, step := range e.spec.Steps {
		s := scope.Scope(fmt.Sprintf("step(%s)", step.Id))

		if exec, err := step.CreateExecutor(ctx, s, e.sExec.JobExecutor(), e.sExec); err != nil {
			return fmt.Errorf("step %q create executor: %w", step.Id, err)
		} else {
			e.children[step.Id] = exec
		}
	}
	return nil
}

func (e *compositeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *compositeActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *compositeActionExecutor) CreateTask(stage Stage) *ActionTask {
	// 1. Plan the execution
	ids := make([]string, len(e.spec.Steps))
	for i, step := range e.spec.Steps {
		ids[i] = step.Id
	}
	if stage == StagePost {
		slices.Reverse(ids) // in-place reverse
	}
	tasks := make([]*StepTask, len(ids))
	for i, id := range ids {
		cExec := e.children[id]
		task := cExec.CreateTask(stage)
		if task != nil {
			task.Run = e.decorator.DecorateStepRun(task)
		}
		tasks[i] = task
	}

	// all runs are nil
	if !slices.ContainsFunc(tasks, func(t *StepTask) bool { return t != nil }) {
		return nil
	}

	// 2. Execute the plan
	run := func(ctx context.Context) error {
		if err := e.computeInputs(); err != nil {
			return err
		}

		for i, task := range tasks {
			if task == nil {
				continue
			}

			id := ids[i]
			res, err := task.Run(ctx)
			if res != nil && res.Conclusion == records.ResultFailure {
				clog.WarnContextf(ctx, "set step.Outcome='failure' because of step %s failed", id)
				e.sExec.SetStatus(records.ResultFailure)
			}
			if err != nil {
				return fmt.Errorf("run step %q (%s): %w", id, stage, err)
			}
		}

		return e.produceOutputs()
	}

	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L472-L490
	var condition workflows.Conditional
	if stage != StageMain {
		condition = "always()"
	}

	return &ActionTask{
		Run:       run,
		Stage:     stage,
		Executor:  e,
		Condition: condition,
	}
}

func (e *compositeActionExecutor) computeInputs() error {
	clear(e.inputs)
	return evaluator.Evaluate(e.exprEnv, e.spec.Inputs, &e.inputs)
}

func (e *compositeActionExecutor) produceOutputs() error {
	outputs := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Outputs, &outputs); err != nil {
		return err
	}
	e.sExec.SetOutput(outputs)
	return nil
}
