package executor

import (
	"context"
	"io"
	"maps"
	"strconv"
	"time"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/pkg/util/dig"

	"go.uber.org/dig"
)

type StepExecutor interface {
	NewChildExecutor(stepRun StepRun) StepExecutor
	ChildExecutor(id string) StepExecutor
	ParentExecutor() StepExecutor
	RootExecutor() StepExecutor

	StepId() string
	Context() context.Context
	SetContext(ctx context.Context)

	Initialize(scope *dig.Scope) error
	RunStep(fn func(StepRun) *Task) *dossiers.Step
	ComposeEnv(m map[string]string)
	SetStatus(status dossiers.Result)

	SetEnv(env map[string]string) error
	CreateStepSummary(r io.Reader) error
	SaveState(state map[string]string) error
	SetOutput(output map[string]string) error
}

func newStepExecutor(job JobExecutor, parent StepExecutor, stepRun StepRun) StepExecutor {
	return &stepExecutor{
		job:      job,
		parent:   parent,
		children: make(map[string]StepExecutor),
		stepRun:  stepRun,
	}
}

type stepExecutor struct {
	job      JobExecutor
	parent   StepExecutor
	children map[string]StepExecutor
	stepRun  StepRun

	// records
	github dossiers.Github
	step   *dossiers.Step
	env    map[string]string
	jobEnv map[string]string
	state  map[string]string // Intra action state

	ctx        context.Context
	exprEnv    expression.Env
	supervisor Supervisor
}

func (e *stepExecutor) Context() context.Context {
	return e.ctx
}

func (e *stepExecutor) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *stepExecutor) StepId() string {
	return e.stepRun.StepId()
}

func (e *stepExecutor) StepRun() StepRun {
	return e.stepRun
}

func (e *stepExecutor) NewChildExecutor(stepRun StepRun) StepExecutor {
	cExec := newStepExecutor(e.job, e, stepRun)
	e.children[stepRun.StepId()] = cExec
	return cExec
}

func (e *stepExecutor) ChildExecutor(id string) StepExecutor {
	return e.children[id]
}

func (e *stepExecutor) ParentExecutor() StepExecutor {
	return e.parent
}

func (e *stepExecutor) RootExecutor() StepExecutor {
	var exec StepExecutor = e
	for exec.ParentExecutor() != nil {
		exec = exec.ParentExecutor()
	}
	return exec
}

func (e *stepExecutor) Initialize(scope *dig.Scope) error {
	// inject dependencies
	if err := xdig.Populate(scope, &e.supervisor); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.github); err != nil {
		return err
	} else {
		e.github.Action = e.StepId()
	}
	if err := xdig.Populate(scope, &e.jobEnv); err != nil {
		return err
	}
	e.env = make(map[string]string)
	maps.Copy(e.env, e.jobEnv)
	e.step = new(dossiers.Step)
	e.step.Outputs = make(map[string]string)
	e.state = make(map[string]string)

	// initialize StepRun
	if err := e.stepRun.Initialize(e, scope); err != nil {
		return err
	}

	if r, ok := e.stepRun.(repository.Repositorial); ok {
		repo := r.Repository()

		e.github.ActionRepository = repo.Name
		e.github.ActionRef = repo.Ref
	}

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("env", e.env),
	}
	if e.parent != nil {
		opts = append(opts, expression.WithLibrary(libraries.StepLib(e.step)))
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

func (e *stepExecutor) RunStep(fn func(StepRun) *Task) *dossiers.Step {
	task := fn(e.stepRun)
	if task == nil {
		return nil
	}

	defer e.endTask(task)
	e.beginTask(task) // TODO logging error
	if e.step.Outcome == "" {
		e.runTask(task)
	}

	return e.step
}

func (e *stepExecutor) beginTask(task *Task) error {
	if err := e.supervisor.BeforeStepRun(e); err != nil {
		return err
	}

	base := e.stepRun.Base()

	clear(e.env)
	maps.Copy(e.env, e.jobEnv)
	if err := evaluator.Evaluate(e.exprEnv, base.Env, &e.env); err != nil {
		e.SetStatus(dossiers.ResultFailure)
		return err
	}

	if meet, err := evaluator.Meet(e.exprEnv, task.Condition); err != nil {
		e.SetStatus(dossiers.ResultFailure)
		e.step.Conclusion = dossiers.ResultFailure
		return err
	} else if !meet {
		e.SetStatus(dossiers.ResultSkipped)
		e.step.Conclusion = dossiers.ResultSkipped
		// TODO logging
	}
	return nil
}

func (e *stepExecutor) runTask(task *Task) error {
	base := e.stepRun.Base()

	timeout := int64(-1)
	if err := evaluator.Evaluate(e.exprEnv, base.TimeoutInMinutes, &timeout); err != nil {
		// TODO logging
		e.SetStatus(dossiers.ResultFailure)
		return err
	} else if timeout > 0 {
		ctx, cancel := context.WithTimeout(e.ctx, time.Duration(timeout)*time.Minute)
		e.ctx = ctx
		defer cancel()
	}

	if err := e.supervisor.BeforeTaskRun(task, e); err != nil {
		// TODO logging
		e.SetStatus(dossiers.ResultFailure)
		return err
	}

	ch := make(chan error)
	go func() {
		ch <- task.Run(e)
	}()

	var err error
	select {
	case <-e.ctx.Done():
		err = e.ctx.Err()
	case err = <-ch:
	}

	if err != nil {
		e.SetStatus(dossiers.ResultFailure)
		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	} else {
		e.SetStatus(dossiers.ResultSuccess)
	}

	if err = e.supervisor.AfterTaskRun(task, e); err != nil {
		// TODO logging
		e.SetStatus(dossiers.ResultFailure)
		return err
	}
	return nil
}

func (e *stepExecutor) endTask(task *Task) {
	if e.step.Outcome == dossiers.ResultFailure {
		base := e.stepRun.Base()

		continueOnError := false
		if err := evaluator.Evaluate(e.exprEnv, base.ContinueOnError, &continueOnError); err != nil {
			//logger.Infof("Failed but continue next step")
			//return err
			e.step.Conclusion = dossiers.ResultFailure
		} else if continueOnError {
			//logger.Infof("Failed but continue next step")
			e.step.Conclusion = dossiers.ResultSuccess
		} else {
			e.step.Conclusion = e.step.Outcome
		}
	}

	if err := e.supervisor.AfterStepRun(e, e.step); err != nil {
		//logger.Infof("Failed but continue next step")
	}
}

func (e *stepExecutor) ComposeEnv(m map[string]string) {
	e.job.ComposeEnv(m)

	maps.Copy(m, e.env)

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
	}
	maps.Copy(m, ghEnv)

	// set STATE_* env
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
	for k, v := range e.state {
		k = "STATE_" + k
		m[k] = v
	}
}

func (e *stepExecutor) SetStatus(status dossiers.Result) {
	e.github.ActionStatus = status
	e.step.Outcome = status
}

// SetEnv make an environment variable available to any subsequent steps in a workflow job
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *stepExecutor) SetEnv(env map[string]string) error {
	for k := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			delete(env, k)
		}
	}
	maps.Copy(e.env, env)
	if e.parent != nil {
		return e.parent.SetEnv(env)
	} else {
		return e.job.SetEnv(env)
	}
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
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) error {
	if e.parent != nil {
		root := e.RootExecutor()
		return root.SaveState(state)
	}
	maps.Copy(e.state, state)
	return nil
}

// SetOutput sets a step's output parameter.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *stepExecutor) SetOutput(output map[string]string) error {
	maps.Copy(e.step.Outputs, output)
	return nil
}
