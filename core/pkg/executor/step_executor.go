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
	"strconv"
	"strings"
	"time"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/expression/libraries"
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
	Status() records.Result // inherit libraries.StatusProvider
	SetStatus(status records.Result)
	Inputs() map[string]string
	SetOutput(output map[string]string)
	Env() map[string]string
	SystemEnv() map[string]string
	SetEnv(env map[string]string)
	SaveState(state map[string]string)
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

type StepRun func(context.Context) (*records.Step, error)

type stepExecutor struct {
	spec   *StepSpec
	parent StepExecutor
	jExec  JobExecutor
	aExec  ActionExecutor

	// records
	github records.Github
	step   *records.Step
	name   string
	inputs map[string]string
	env    map[string]string
	state  map[string]string // Intra action state

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
		fmt.Sprintf("StepExecutor.Initialize(%s)", stepId),
		xotel.DrassiStep(stepId),
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
	e.github.Action = e.spec.Id
	e.env = maps.Clone(e.upperEnv())
	e.step = new(records.Step)
	e.step.Outputs = make(map[string]string)
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
			expression.WithLibrary(libraries.StatusLib(e.parent)),
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
	if err := e.evaluateDisplayName(s); err != nil {
		return err
	}

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

	run := func(ctx context.Context) (*records.Step, error) {
		err := e.runAction(ctx, task)
		if e.step.Outcome == "" {
			if err != nil {
				e.SetStatus(records.ResultFailure)
			} else {
				e.SetStatus(records.ResultSuccess)
			}
		}
		if e.step.Conclusion == "" {
			e.step.Conclusion = e.step.Outcome
		}
		return e.step, err
	}
	return &StepTask{
		Run:      run,
		Stage:    stage,
		Executor: e,
	}
}

func (e *stepExecutor) telemetry(stage Stage, run ActionRun) ActionRun {
	return func(ctx context.Context) (err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("ActionRun(%s, %s)", e.spec.Id, stage),
		)
		defer done(&err)

		return run(ctx)
	}
}

func (e *stepExecutor) runAction(ctx context.Context, action *ActionTask) error {
	s := scribe.FromContext(ctx)

	maps.Copy(e.env, e.upperEnv())
	s.Debugf("Evaluating 'env' for step: (%s)", e.spec.Id)
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Env(), e.spec.Env), &e.env); err != nil {
		return fmt.Errorf("evaluate 'env': %w", err)
	}

	if err := e.evaluateDisplayName(s); err != nil {
		return err
	}
	displayName := e.Name(action.Stage)

	s.Debugf("Evaluating 'condition' for step: %q (%s)", displayName, e.spec.Id)
	if meet, err := evaluator.Meet(e.exprEnv, action.Condition); err != nil {
		s.Errorf("Error while evaluate 'if': %v", err)
		return fmt.Errorf("evaluate 'if': %w", err)
	} else if !meet {
		e.SetStatus(records.ResultSkipped)
		s.Writef("Skipped step %q (%s)", displayName, e.spec.Id)
		return nil
	}

	timeout := int64(-1)
	s.Debugf("Evaluating 'timeout-minutes' for step: %q (%s)", displayName, e.spec.Id)
	if err := evaluator.Evaluate(e.exprEnv, e.spec.TimeoutInMinutes, &timeout); err != nil {
		s.Errorf("Error while evaluate 'timeout-minutes': %v", err)
		return fmt.Errorf("evaluate 'timeout-minutes': %w", err)
	} else if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
		defer cancel()
	}

	clear(e.inputs)
	s.Debugf("Evaluating 'inputs' for step: %q (%s)", displayName, e.spec.Id)
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Inputs(), e.spec.Inputs), &e.inputs); err != nil {
		return fmt.Errorf("evaluate 'inputs': %w", err)
	}

	err := action.Run(ctx)
	if err != nil {
		s.Errorf("Error while running task %q (%s): %v", displayName, e.spec.Id, err)
		e.SetStatus(records.ResultFailure)
	}

	// NOTE: step.Outcome can be set from outside StepExecutor, e.g: CompositeAction or CommandProcessor
	if e.step.Outcome == records.ResultFailure {
		continueOnError := false
		s.Debugf("Evaluating 'continue-on-error' for step: %q (%s)", displayName, e.spec.Id)
		if err := evaluator.Evaluate(e.exprEnv, e.spec.ContinueOnError, &continueOnError); err != nil {
			s.Errorf("Error while evaluate 'continue-on-error': %v", err)
			return fmt.Errorf("evaluate 'continue-on-error': %w", err)
		} else if continueOnError {
			e.step.Conclusion = records.ResultSuccess
			s.Warningf("Step failed but continue next step")
			return nil
		}
	}

	outputs := make(map[string]string)
	s.Debugf("Evaluating 'outputs' for step: %q (%s)", displayName, e.spec.Id)
	if err := evaluator.Evaluate(e.exprEnv, mergeMapExpr(e.aExec.Outputs(), e.spec.Outputs), &outputs); err != nil {
		return fmt.Errorf("evaluate 'outputs': %w", err)
	}
	e.SetOutput(outputs)

	return err
}

func (e *stepExecutor) evaluateDisplayName(s *scribe.Scribe) error {
	name, prefix := "", ""
	expr := e.spec.Name
	if expr == nil {
		expr = e.aExec.Name()
		prefix = "Run "
	}

	s.Debugf("Evaluating display name")
	if err := evaluator.Evaluate(e.exprEnv, expr, &name); err != nil {
		return fmt.Errorf("evaluate 'name': %w", err)
	}

	name = strings.TrimLeft(name, " \t\r\n")
	name, _, _ = strings.Cut(name, "\n")
	name = strings.TrimSpace(name)
	name = prefix + name

	s.Debugf("Set step %q display name to: %q", e.spec.Id, name)
	e.name = name
	return nil
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

// Status return current step's Outcome
// its implement [libraries.StatusProvider]
func (e *stepExecutor) Status() records.Result {
	if e.step != nil {
		return e.step.Outcome
	}
	return records.ResultSuccess
}

func (e *stepExecutor) SetStatus(status records.Result) {
	e.github.ActionStatus = status
	e.step.Outcome = status
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
	maps.Copy(e.step.Outputs, output)
}

// Env return environment variable
func (e *stepExecutor) Env() map[string]string {
	return e.env
}

func (e *stepExecutor) SystemEnv() map[string]string {
	m := e.envProv.Env(e)

	jEnv := e.jExec.SystemEnv()
	maps.Copy(m, jEnv)

	// set GITHUB_* env
	ghEnv := map[string]string{
		"GITHUB_ACTION":              e.github.Action,
		"GITHUB_ACTION_REF":          e.github.ActionRef,
		"GITHUB_ACTION_REPOSITORY":   e.github.ActionRepository,
		"GITHUB_ACTOR":               e.github.Actor,
		"GITHUB_ACTOR_ID":            e.github.ActorId,
		"GITHUB_API_URL":             e.github.ApiUrl,
		"GITHUB_BASE_REF":            e.github.BaseRef,
		"GITHUB_EVENT_NAME":          e.github.EventName,
		"GITHUB_EVENT_PATH":          e.github.EventPath,
		"GITHUB_GRAPHQL_URL":         e.github.GraphqlUrl,
		"GITHUB_HEAD_REF":            e.github.HeadRef,
		"GITHUB_JOB":                 e.github.Job,
		"GITHUB_REF":                 e.github.Ref,
		"GITHUB_REF_NAME":            e.github.RefName,
		"GITHUB_REF_PROTECTED":       strconv.FormatBool(e.github.RefProtected),
		"GITHUB_REF_TYPE":            string(e.github.RefType),
		"GITHUB_REPOSITORY":          e.github.Repository,
		"GITHUB_REPOSITORY_ID":       e.github.RepositoryId,
		"GITHUB_REPOSITORY_OWNER":    e.github.RepositoryOwner,
		"GITHUB_REPOSITORY_OWNER_ID": e.github.RepositoryOwnerId,
		"GITHUB_RETENTION_DAYS":      e.github.RetentionDays,
		"GITHUB_RUN_ATTEMPT":         e.github.RunAttempt,
		"GITHUB_RUN_ID":              e.github.RunId,
		"GITHUB_RUN_NUMBER":          e.github.RunNumber,
		"GITHUB_SERVER_URL":          e.github.ServerUrl,
		"GITHUB_SHA":                 e.github.Sha,
		"GITHUB_TRIGGERING_ACTOR":    e.github.TriggeringActor,
		"GITHUB_WORKFLOW":            e.github.Workflow,
		"GITHUB_WORKFLOW_REF":        e.github.WorkflowRef,
		"GITHUB_WORKFLOW_SHA":        e.github.WorkflowSha,
		"GITHUB_WORKSPACE":           e.github.Workspace,
	}
	maps.Copy(m, ghEnv)

	// set STATE_* env
	// https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-commands#sending-values-to-the-pre-and-post-actions
	for k, v := range e.state {
		k = "STATE_" + k
		m[k] = v
	}
	return m
}

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
// Environment variables should be applied to all StepExecutor in the Stack as well as the JobExecutor.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *stepExecutor) SetEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

// SaveState used to create environment variables for sharing pre: or post: action state.
// You should identify the root step by using [Root] to store the state.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) {
	maps.Copy(e.state, state)
}
