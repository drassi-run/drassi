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

type CompositeActionSpec struct {
	Inputs  workflows.Evaluable[map[string]string]
	Outputs workflows.Evaluable[map[string]string]

	Steps []*StepSpec
}

func (spec *CompositeActionSpec) CreateExecutor(ctx context.Context, scope *dig.Scope) (ActionExecutor, error) {
	e := &compositeActionExecutor{spec: spec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type compositeActionExecutor struct {
	spec *CompositeActionSpec

	decorator StepRunDecorator
	exprEnv   expression.Env
	inputs    map[string]string
	children  map[string]StepExecutor
}

func (e *compositeActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	var exec StepExecutor
	if err := xdig.Populate(scope, &exec); err != nil {
		return err
	}
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
		expression.WithLibrary(libraries.StatusLib(exec)),
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
		cScope := scope.Scope(fmt.Sprintf("step(%s)", step.Id))

		if cExec, err := step.CreateExecutor(ctx, cScope); err != nil {
			return err
		} else {
			e.children[step.Id] = cExec
		}
	}
	return nil
}

func (e *compositeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *compositeActionExecutor) CreateRun(stage Stage) *ActionRun {
	// 1. Plan the execution
	ids := make([]string, len(e.spec.Steps))
	for i, step := range e.spec.Steps {
		ids[i] = step.Id
	}
	if stage == StagePost {
		slices.Reverse(ids) // in-place reverse
	}
	runs := make([]StepRun, len(ids))
	for i, id := range ids {
		cExec := e.children[id]
		run := cExec.CreateRun(stage)
		if run != nil {
			run = e.decorator.DecorateStepRun(stage, cExec, run)
		}
		runs[i] = run
	}

	// all runs are nil
	if !slices.ContainsFunc(runs, func(r StepRun) bool { return r != nil }) {
		return nil
	}

	// 2. Execute the plan
	taskRun := func(ctx context.Context, exec StepExecutor) error {
		if err := e.computeInputs(); err != nil {
			return err
		}

		for i, run := range runs {
			if run == nil {
				continue
			}
			res := run(ctx)
			if res != nil && res.Conclusion == records.ResultFailure {
				exec.SetStatus(records.ResultFailure)
				return fmt.Errorf("step %q (%s) failed", ids[i], stage)
			}
		}

		return e.produceOutputs(exec)
	}

	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L472-L490
	var condition workflows.Conditional
	if stage != StageMain {
		condition = "always()"
	}

	return &ActionRun{
		Condition: condition,
		Run:       taskRun,
	}
}

func (e *compositeActionExecutor) computeInputs() error {
	clear(e.inputs)
	return evaluator.Evaluate(e.exprEnv, e.spec.Inputs, &e.inputs)
}

func (e *compositeActionExecutor) produceOutputs(exec StepExecutor) error {
	outputs := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Outputs, &outputs); err != nil {
		return err
	}
	exec.SetOutput(outputs)
	return nil
}
