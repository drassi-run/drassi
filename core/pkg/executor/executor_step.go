/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/util/dig"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
)

type StepExecutor interface {
	StepRun() StepRun

	Initialize(ctx context.Context, scope *dig.Scope) error
	RunStep(ctx context.Context, stage Stage) *records.Step

	Status() records.Result
	SetStatus(status records.Result)
	ComposeEnv(systemEnv bool) map[string]string
	SetEnv(env map[string]string)
	SaveState(state map[string]string)
	SetOutput(output map[string]string)
	CreateStepSummary(r io.Reader) error
}

func StepId(e StepExecutor) string {
	return e.StepRun().StepId()
}

func StepUid(e StepExecutor) string {
	return e.StepRun().Base().Uid
}

func NewStepExecutor(stepRun StepRun) StepExecutor {
	exec := &stepExecutor{stepRun: stepRun}
	return WithTelemetryStepExecutor(exec)
}

type stepExecutor struct {
	stepRun StepRun

	// records
	github   records.Github
	step     *records.Step
	env      map[string]string
	upperEnv map[string]string // env variables from upper layers
	state    map[string]string // Intra action state

	tracker  Tracker
	listener StepListener
	exprEnv  expression.Env
}

func (e *stepExecutor) StepRun() StepRun {
	return e.stepRun
}

func (e *stepExecutor) Initialize(ctx context.Context, scope *dig.Scope) (ex error) {
	l := clog.FromContext(ctx)

	// setup listener
	if err := xdig.Populate(scope, &e.listener); err != nil {
		l.Errorf("Failed to initialize step: %v", err)
		return err
	}
	if eh := e.listener.OnInitializeStep(e, scope); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &ex)
		if err != nil {
			return err
		}
	}

	// do step initialization
	// inject dependencies
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.github); err != nil {
		return err
	} else {
		e.github.Action = StepId(e)
	}
	if err := xdig.Populate(scope, &e.upperEnv); err != nil {
		return err
	}
	e.env = make(map[string]string)
	maps.Copy(e.env, e.upperEnv)
	e.step = new(records.Step)
	e.step.Outputs = make(map[string]string)
	e.state = make(map[string]string)

	// initialize StepRun
	if err := xdig.Supply[StepExecutor](scope, e); err != nil {
		return err
	}
	if err := e.stepRun.Initialize(ctx, scope); err != nil {
		return err
	}

	if r, ok := e.stepRun.(interface{ Repository() *repository.Repository }); ok {
		repo := r.Repository()

		e.github.ActionRepository = repo.Name
		e.github.ActionRef = repo.Ref
	}

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("env", e.env),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
		return err
	} else {
		e.exprEnv = exprEnv
	}

	// Provide scope values
	if err := xdig.Supply(scope, e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.github); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.env); err != nil {
		return err
	}

	return nil
}

func (e *stepExecutor) RunStep(ctx context.Context, stage Stage) *records.Step {
	task := e.getTask(stage)
	if task == nil {
		return nil
	}

	// setup listener
	if eh := e.listener.OnRunStep(e, stage); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &err)
		if err != nil {
			return nil
		}
	}

	// do step run
	e.runTask(ctx, task)
	return e.step
}

func (e *stepExecutor) getTask(stage Stage) *Task {
	switch stage {
	case StagePre:
		return e.stepRun.PreTask()
	case StageMain:
		return e.stepRun.MainTask()
	case StagePost:
		return e.stepRun.PostTask()
	default:
		return nil
	}
}

func (e *stepExecutor) runTask(ctx context.Context, task *Task) {
	s := scribe.FromContext(ctx)
	base := e.stepRun.Base()
	stepId, displayName := base.StepId(), e.stepRun.DisplayName(task.Stage)

	clear(e.env)
	maps.Copy(e.env, e.upperEnv)
	if err := evaluator.Evaluate(e.exprEnv, base.Env, &e.env); err != nil {
		e.SetStatus(records.ResultFailure)
		return
	}

	s.Debugf("Evaluating condition for step: %q (%s)", displayName, stepId)
	if meet, err := evaluator.Meet(e.exprEnv, task.Condition); err != nil {
		e.SetStatus(records.ResultFailure)
		e.step.Conclusion = records.ResultFailure
		s.Errorf("Error while evaluate 'if': %v", err)
		return
	} else if !meet {
		e.SetStatus(records.ResultSkipped)
		e.step.Conclusion = records.ResultSkipped
		s.Writef("Skipped step %q (%s)", displayName, stepId)
		return
	}

	timeout := int64(-1)
	s.Debugf("Evaluating 'timeout-minutes' for step: %q (%s)", displayName, stepId)
	if err := evaluator.Evaluate(e.exprEnv, base.TimeoutInMinutes, &timeout); err != nil {
		s.Errorf("Error while evaluate 'timeout-minutes': %v", err)
		e.SetStatus(records.ResultFailure)
		return
	} else if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
		defer cancel()

		stop := e.tracker.StartContext(ctx)
		defer stop()
	}

	if err := e.doTask(ctx, task); err != nil {
		s.Errorf("Error while running task %q (%s): %v", displayName, stepId, err)
		e.SetStatus(records.ResultFailure)
	} else {
		e.SetStatus(records.ResultSuccess)
	}

	if e.step.Outcome == records.ResultFailure {
		continueOnError := false
		s.Debugf("Evaluating 'continue-on-error' for step: %q (%s)", displayName, stepId)
		if err := evaluator.Evaluate(e.exprEnv, base.ContinueOnError, &continueOnError); err != nil {
			e.step.Conclusion = records.ResultFailure
			s.Errorf("Error while evaluate 'continue-on-error' %v", err)
		} else if continueOnError {
			e.step.Conclusion = records.ResultSuccess
			s.Warningf("Step failed but continue next step")
		}
	}
	if e.step.Conclusion == "" {
		e.step.Conclusion = e.step.Outcome
	}
}

func (e *stepExecutor) doTask(ctx context.Context, task *Task) (ex error) {
	// setup listener
	if eh := e.listener.OnRunTask(e, task); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &ex)
		if err != nil {
			return err
		}
	}

	return task.Run(ctx, e)
}

func (e *stepExecutor) ComposeEnv(systemEnv bool) map[string]string {
	m := maps.Clone(e.env)
	if !systemEnv {
		return m
	}

	maps.Copy(m, e.tracker.Env())

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
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
	for k, v := range e.state {
		k = "STATE_" + k
		m[k] = v
	}
	return m
}

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

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
// Environment variables should be applied to all StepExecutor in the Stack as well as the JobExecutor.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *stepExecutor) SetEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

// CreateStepSummary create custom Markdown that it will be displayed on the summary page of a workflow run.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (e *stepExecutor) CreateStepSummary(io.Reader) error {
	//TODO implement me
	return nil
}

// SaveState used to create environment variables for sharing pre: or post: action state.
// You should identify the root step by using [Stack.Root] to store the state.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) {
	maps.Copy(e.state, state)
}

// SetOutput sets a step's output parameter.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *stepExecutor) SetOutput(output map[string]string) {
	maps.Copy(e.step.Outputs, output)
}
