/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.uber.org/dig"
)

type StepExecutor interface {
	StepSpec() *StepSpec
	Parent() StepExecutor
	JobExecutor() JobExecutor
	ActionExecutor() ActionExecutor
	CreateTask(stage Stage) *StepTask

	ExprEnv() expression.Env
	Sandbox() sandboxer.Sandbox
	Streams(ctx context.Context, stage Stage) *stream.Streams

	Name(stage Stage) string
	Github() *records.Github
	Inputs() map[string]string
	SetOutput(output map[string]string)
	Env() map[string]string
	SetEnv(env map[string]string)
	State() map[string]string
	SaveState(state map[string]string)

	ComposeEnv() map[string]string
}

func Root(exec StepExecutor) StepExecutor {
	for exec.Parent() != nil {
		exec = exec.Parent()
	}
	return exec
}

func Depth(exec StepExecutor) (d int) {
	for d = 1; exec.Parent() != nil; d++ {
		exec = exec.Parent()
	}
	return
}

type StepTask struct {
	Run      StepRun
	Stage    Stage
	Executor StepExecutor
}

func (t *StepTask) StepId() string {
	return t.StepSpec().Id
}

func (t *StepTask) StepSpec() *StepSpec {
	return t.Executor.StepSpec()
}

func (t *StepTask) JobSpec() *JobSpec {
	return t.Executor.JobExecutor().JobSpec()
}

type StepRun func(context.Context) (*records.StepResult, error)

type stepExecutor struct {
	spec   *StepSpec
	parent StepExecutor
	jExec  JobExecutor
	aExec  ActionExecutor

	// records
	github *records.Github

	name    string
	inputs  map[string]string
	outputs map[string]string
	env     map[string]string
	state   map[string]string // Intra action state

	decorator ActionRunDecorator
	envProv   EnvProvider
	exprEnv   expression.Env
	factory   stream.Factory[Milieu]
}

func (e *stepExecutor) StepSpec() *StepSpec {
	return e.spec
}

func (e *stepExecutor) Parent() StepExecutor {
	return e.parent
}

func (e *stepExecutor) JobExecutor() JobExecutor {
	return e.jExec
}

func (e *stepExecutor) ActionExecutor() ActionExecutor {
	return e.aExec
}

func (e *stepExecutor) init(ctx context.Context, scope *dig.Scope) (ex error) {
	stepId := e.spec.Id
	ctx, done := xotel.SetupTelemetry(ctx,
		fmt.Sprintf("StepExecutor.init(%s)", stepId),
		xotel.Step("init/"+stepId),
	)
	defer done(&ex)

	// do step initialization
	// inject dependencies
	if err := xdig.Populate(scope, &e.decorator); err != nil {
		return fmt.Errorf("populate 'decorator': %w", err)
	}
	if err := xdig.Populate(scope, &e.envProv); err != nil {
		return fmt.Errorf("populate 'envProv': %w", err)
	}
	if err := xdig.Populate(scope, &e.factory); err != nil {
		return fmt.Errorf("populate 'stream.Factory': %w", err)
	}
	if err := xdig.Populate(scope, &e.github); err != nil {
		return fmt.Errorf("populate 'github': %w", err)
	}
	e.github = new(*e.github) // clone
	e.github.Action = e.spec.Id
	e.inputs = make(map[string]string)
	e.outputs = make(map[string]string)
	e.env = maps.Clone(e.upperEnv())
	e.state = make(map[string]string)

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("env", e.env),
	}
	if e.parent == nil {
		e.exprEnv = e.jExec.ExprEnv()
	} else {
		e.exprEnv = e.parent.ExprEnv()
		opts = append(opts,
			// inputs from upper layers will NOT be passed to child steps
			expression.WithVariable("inputs", e.inputs),
		)
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
		return fmt.Errorf("create child expression.Env: %w", err)
	} else {
		e.exprEnv = exprEnv
	}

	// initialize StepRun
	if exec, err := e.spec.Action.CreateExecutor(ctx, scope, e); err != nil {
		return fmt.Errorf("create ActionExecutor for %q: %w", e.spec.Id, err)
	} else {
		e.aExec = exec
	}

	if r, ok := e.spec.Action.(interface{ Repository() *repository.Repository }); ok {
		repo := r.Repository()

		e.github.ActionRepository = repo.Name
		e.github.ActionRef = repo.Ref
	}

	// initialize displayName
	s := scribe.FromContext(ctx)
	e.tryUpdateDisplayName(s)

	// Provide scope values
	if err := xdig.Supply(scope, e.exprEnv); err != nil {
		return fmt.Errorf("supply 'exprEnv': %w", err)
	}
	if err := xdig.Supply(scope, e.github); err != nil {
		return fmt.Errorf("supply 'github': %w", err)
	}
	if err := xdig.Supply(scope, e.env); err != nil {
		return fmt.Errorf("supply 'env': %w", err)
	}

	return nil
}

func (e *stepExecutor) CreateTask(stage Stage) *StepTask {
	task := e.aExec.CreateTask(stage)
	if task == nil {
		return nil
	}
	if stage == StageMain {
		task.Condition = e.spec.Condition
	}
	task.Run = e.decorator.DecorateActionRun(task)
	task.Run = e.telemetry(stage, task.Run)

	run := func(ctx context.Context) (*records.StepResult, error) {
		clear(e.inputs)
		clear(e.outputs)
		res := &records.StepResult{
			Outcome: records.ResultSuccess,
			Outputs: make(map[string]string),
		}

		err := e.runAction(ctx, task, res)

		if res.Outcome == "" {
			if err != nil {
				res.Outcome = records.ResultFailure
			} else {
				res.Outcome = records.ResultSuccess
			}
		}

		res.Conclusion = res.Outcome
		if res.Outcome == records.ResultFailure {
			continueOnError := false
			s := scribe.FromContext(ctx)
			s.Debugf("Evaluating 'continue-on-error'")
			if err := evaluator.Evaluate(e.exprEnv, e.spec.ContinueOnError, &continueOnError); err != nil {
				s.Errorf("Evaluate 'continue-on-error' error: %v", err)
			}
			if continueOnError {
				s.Warningf("Step failed but continue next step")
				res.Conclusion = records.ResultSuccess
			}
		}

		maps.Copy(res.Outputs, e.outputs)
		return res, err
	}

	return &StepTask{
		Run:      run,
		Stage:    stage,
		Executor: e,
	}
}

func (e *stepExecutor) telemetry(stage Stage, run ActionRun) ActionRun {
	return func(ctx context.Context) (_ records.Result, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("ActionRun(%s/%s)", stage, e.spec.Id),
		)
		defer done(&err)

		return run(ctx)
	}
}

func (e *stepExecutor) runAction(ctx context.Context, action *ActionTask, sr *records.StepResult) error {
	s := scribe.FromContext(ctx)

	maps.Copy(e.env, e.upperEnv())
	s.Debugf("Evaluating 'env'")
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Env(), e.spec.Env), &e.env); err != nil {
		s.Errorf("Evaluate 'env' error: %v", err)
		return fmt.Errorf("evaluate 'env': %w", err)
	}

	e.tryUpdateDisplayName(s)
	displayName := e.Name(action.Stage)

	s.Debugf("Evaluating 'if'")
	if meet, err := evaluator.Meet(e.exprEnv, action.Condition); err != nil {
		// condition fail also bypass `continue-on-error`
		s.Errorf("Evaluate 'if' error: %v", err)
		return fmt.Errorf("evaluate 'if': %w", err)
	} else if !meet {
		sr.Outcome = records.ResultSkipped
		s.Writef("Skipped step %q (%s)", displayName, e.spec.Id)
		return nil
	}

	timeout := int64(-1)
	s.Debugf("Evaluating 'timeout-minutes'")
	if err := evaluator.Evaluate(e.exprEnv, e.spec.TimeoutInMinutes, &timeout); err != nil {
		// the error is logged but the step continues with no timeout set.
		s.Errorf("Evaluate 'timeout-minutes' error: %v", err)
	} else if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
		defer cancel()
	}

	clear(e.inputs)
	s.Debugf("Evaluating 'inputs'")
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Inputs(), e.spec.Inputs), &e.inputs); err != nil {
		s.Errorf("Evaluate 'inputs' error: %v", err)
		return fmt.Errorf("evaluate 'inputs': %w", err)
	}

	res, err := action.Run(ctx)
	switch res {
	case records.ResultCancelled:
		if err != nil {
			s.Errorf("The operation was canceled: %v", err)
		} else {
			s.Errorf("The operation was canceled")
		}
	case records.ResultFailure:
		s.Errorf("Running task error: %v", err)
	}
	sr.Outcome = res

	outputs := make(map[string]string)
	s.Debugf("Evaluating 'outputs'")
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Outputs(), e.spec.Outputs), &outputs); err != nil {
		sr.Outcome = records.ResultFailure
		s.Errorf("Evaluate 'outputs' error: %v", err)
		return fmt.Errorf("evaluate 'outputs': %w", err)
	}
	maps.Copy(sr.Outputs, outputs)

	return err
}

func (e *stepExecutor) tryUpdateDisplayName(s *scribe.Scribe) {
	name, prefix := "", ""
	expr := e.spec.Name
	if expr == nil {
		expr = e.aExec.Name()
		prefix = "Run "
	}

	s.Debugf("Evaluating display name")
	if err := evaluator.Evaluate(e.exprEnv, expr, &name); err != nil {
		s.Errorf("Evaluate 'name' error: %v", err)
		return
	}

	name = strings.TrimLeft(name, " \t\r\n")
	name, _, _ = strings.Cut(name, "\n")
	name = strings.TrimSpace(name)
	name = prefix + name

	s.Debugf("Set step %q display name to: %q", e.spec.Id, name)
	e.name = name
}

func (e *stepExecutor) upperEnv() map[string]string {
	if e.parent != nil {
		return e.parent.Env()
	}
	return e.JobExecutor().Env()
}

func (e *stepExecutor) ExprEnv() expression.Env {
	return e.exprEnv
}

func (e *stepExecutor) Sandbox() sandboxer.Sandbox {
	return e.jExec.Sandbox()
}

func (e *stepExecutor) Streams(ctx context.Context, stage Stage) *stream.Streams {
	s := NewMilieu(stage, e)
	return e.factory.Create(ctx, s)
}

func (e *stepExecutor) Name(stage Stage) string {
	switch stage {
	case StagePre:
		return "Pre " + e.name
	case StagePost:
		return "Post " + e.name
	default:
		return e.name
	}
}

func (e *stepExecutor) Github() *records.Github {
	return e.github
}

// Inputs return evaluated inputs
func (e *stepExecutor) Inputs() map[string]string {
	return e.inputs
}

// SetOutput sets a step's output parameter.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *stepExecutor) SetOutput(output map[string]string) {
	maps.Copy(e.outputs, output)
}

// Env return environment variable
func (e *stepExecutor) Env() map[string]string {
	return e.env
}

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
// Environment variables should be applied to all StepExecutor in the Stack as well as the JobExecutor.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *stepExecutor) SetEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

func (e *stepExecutor) ComposeEnv() map[string]string {
	m := e.envProv.Env(e)
	maps.Copy(m, e.Env())
	return m
}

// State return intra action state
func (e *stepExecutor) State() map[string]string {
	return e.state
}

// SaveState used to create environment variables for sharing pre: or post: action state.
// You should identify the root step by using [Root] to store the state.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) {
	maps.Copy(e.state, state)
}
