/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.uber.org/dig"
)

type CompositeActionSpec struct {
	Repo    *repository.Repository
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
	children  map[string]StepExecutor
	status    records.Result
}

func (e *compositeActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.decorator); err != nil {
		return err
	}

	e.status = records.ResultSuccess
	if err := scope.Decorate(e.overrideExpressionEnv); err != nil {
		return fmt.Errorf("override StatusLib in expression.Env: %w", err)
	}

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

func (e *compositeActionExecutor) overrideExpressionEnv(exprEnv expression.Env) (expression.Env, error) {
	opt := expression.WithLibrary(libraries.StatusLib(e))
	return exprEnv.New(opt)
}

func (e *compositeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *compositeActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *compositeActionExecutor) Name() workflows.Evaluable[string] {
	name := repository.Location(e.spec.Repo)
	return workflows.NewLiteralToken(name)
}

func (e *compositeActionExecutor) Env() workflows.Evaluable[map[string]string] {
	return nil
}

func (e *compositeActionExecutor) Inputs() workflows.Evaluable[map[string]string] {
	return e.spec.Inputs
}

func (e *compositeActionExecutor) Outputs() workflows.Evaluable[map[string]string] {
	return e.spec.Outputs
}

func (e *compositeActionExecutor) CreateTask(stage Stage) *ActionTask {
	ids := make([]string, len(e.spec.Steps))
	for i, step := range e.spec.Steps {
		ids[i] = step.Id
	}
	if stage == StagePost {
		slices.Reverse(ids) // in-place reverse
	}

	tasks := make([]*StepTask, 0)
	for _, id := range ids {
		cExec := e.children[id]
		task := cExec.CreateTask(stage)
		if task == nil {
			continue
		}
		task.Run = e.decorator.DecorateStepRun(task)
		task.Run = e.telemetry(stage, id, task.Run)
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return nil
	}

	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L472-L490
	var condition workflows.Conditional
	if stage != StageMain {
		condition = "always()"
	}

	return &ActionTask{
		Run:       e.runStage(stage, tasks),
		Stage:     stage,
		Executor:  e,
		Condition: condition,
	}
}

func (e *compositeActionExecutor) runStage(stage Stage, tasks []*StepTask) ActionRun {
	return func(ctx context.Context) (records.Result, error) {
		forge := e.sExec.Forge()
		errs := make([]error, 0)
		cause := error(nil)
		e.status = records.ResultSuccess

		for _, task := range tasks {
			forge.ActionStatus = e.status

			res, err := task.Run(ctx)
			if con := res.Conclusion; level(e.status) < level(con) {
				e.status = con
			}

			if err != nil {
				stepId := task.StepId()
				switch e.status {
				case records.ResultFailure:
					err = fmt.Errorf("step %s/%s fail: %w", stage, stepId, err)
					errs = append(errs, err)
				case records.ResultCancelled:
					if cause == nil {
						// only record first cause
						cause = fmt.Errorf("step %s/%s canceled: %w", stage, stepId, err)
					}
				}
			}
		}

		switch result := e.status; result {
		case records.ResultFailure:
			return result, errors.Join(errs...)
		case records.ResultCancelled:
			return result, cause
		default:
			return result, nil
		}
	}
}

func (e *compositeActionExecutor) telemetry(stage Stage, stepId string, run StepRun) StepRun {
	return func(ctx context.Context) (_ *records.StepResult, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s/%s)", stage, stepId),
			xotel.Step(string(stage)+"/"+stepId),
		)
		defer done(&err)

		return run(ctx)
	}
}

// Status return current step's Outcome
// its implement [libraries.StatusProvider]
func (e *compositeActionExecutor) Status() records.Result {
	return e.status
}
