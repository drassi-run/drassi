package executor

import (
	"context"
	"maps"
	"time"

	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/sandboxer"
)

type StepCommandHandler interface {
	StepRun() StepRun
	Sandbox() sandboxer.Sandbox

	SetEnv(env map[string]string, inExprValue bool) error
	CreateStepSummary() error
	SaveState(state map[string]string) error
	SetOutput(output map[string]string) error
}

type StepExecutor interface {
	JobExecutor() JobExecutor
	NewChildExecutor(stepRun StepRun) StepExecutor
	ChildExecutor(id string) StepExecutor
	ParentExecutor() StepExecutor
	RootExecutor() StepExecutor

	StepId() string
	Streams() *sandboxer.Streams
	Sandbox() sandboxer.Sandbox
	Dossier() *dossiers.Dossier

	Initialize(ctx context.Context) error
	RunStep(ctx context.Context, fn func(StepRun) *Task) *dossiers.Step
	ComposeEnv() map[string]string
	SetResult(outcome dossiers.Result)
}

type stepExecutor struct {
	job      JobExecutor
	parent   StepExecutor
	children map[string]StepExecutor
	stepRun  StepRun
	reporter reporter.Reporter
	cmdCtrl  CommandController

	// Intra action state
	state map[string]string

	dossier  *dossiers.Dossier
	result   *dossiers.Step
	extraEnv map[string]string
}

func (e *stepExecutor) StepId() string {
	return e.stepRun.StepId()
}

func (e *stepExecutor) StepRun() StepRun {
	return e.stepRun
}

func (e *stepExecutor) JobExecutor() JobExecutor {
	return e.job
}

func (e *stepExecutor) NewChildExecutor(stepRun StepRun) StepExecutor {
	cExec := &stepExecutor{
		job:      e.job,
		parent:   e,
		children: make(map[string]StepExecutor),
		stepRun:  stepRun,
		reporter: e.reporter,
		cmdCtrl:  e.cmdCtrl,
		state:    make(map[string]string),
	}

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

func (e *stepExecutor) Streams() *sandboxer.Streams {
	return e.job.Streams()
}

func (e *stepExecutor) Sandbox() sandboxer.Sandbox {
	return e.job.Sandbox()
}

func (e *stepExecutor) Dossier() *dossiers.Dossier {
	return e.dossier
}

func (e *stepExecutor) Initialize(ctx context.Context) error {
	return e.stepRun.Initialize(ctx, e)
}

func (e *stepExecutor) RunStep(ctx context.Context, fn func(StepRun) *Task) *dossiers.Step {
	task := fn(e.stepRun)
	if task == nil {
		return nil
	}

	e.initTask()
	defer e.endTask(ctx, task)
	e.beginTask(ctx, task)
	if e.result.Outcome == "" {
		e.runTask(ctx, task)
	}

	return e.result
}

func (e *stepExecutor) initTask() {
	e.dossier = e.job.NewSubDossier()
	e.stepRun.SetContextInfo(e.dossier)
	e.result = &dossiers.Step{
		Outputs: make(map[string]string),
	}
	e.extraEnv = make(map[string]string)
}

func (e *stepExecutor) beginTask(ctx context.Context, task *Task) {
	if e.parent == nil { // root step
		e.reporter.StartStep(e.StepId())
	}

	base := e.stepRun.Base()
	evalSupplier := &evaluationSupplier{dossier: e.dossier}

	if env, err := base.Env.Evaluate("job.step.env", evalSupplier); err != nil {
		// TODO logging
		e.result.Outcome = dossiers.ResultFailure
		return
	} else {
		maps.Copy(e.dossier.Env, env)
	}

	if task.Condition != nil {
		if meet, err := task.Condition.Meet("job.step", evalSupplier); err != nil {
			e.result.Conclusion = dossiers.ResultFailure
			e.result.Outcome = dossiers.ResultFailure
		} else if !meet {
			e.result.Conclusion = dossiers.ResultSkipped
			e.result.Outcome = dossiers.ResultSkipped
			// TODO logging
		}
	}
}

func (e *stepExecutor) runTask(ctx context.Context, task *Task) {
	base := e.stepRun.Base()
	evalSupplier := &evaluationSupplier{dossier: e.dossier}

	if !base.TimeoutInMinutes.IsNil() {
		if timeout, err := base.TimeoutInMinutes.Evaluate("job.step.timeout-minutes", evalSupplier); err != nil {
			// TODO logging
			e.result.Outcome = dossiers.ResultFailure
			return
		} else if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
			defer cancel()
		}
	}

	if env, err := e.cmdCtrl.StartStep(ctx, e); err != nil {
		// TODO logging
		e.result.Outcome = dossiers.ResultFailure
		return
	} else {
		maps.Copy(e.extraEnv, env)
	}

	ch := make(chan error)
	go func() {
		ch <- task.Run(ctx, e)
	}()

	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}

	if err != nil {
		e.result.Outcome = dossiers.ResultFailure
		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	} else {
		e.result.Conclusion = dossiers.ResultSuccess
		e.result.Outcome = dossiers.ResultSuccess
	}

	if err = e.cmdCtrl.EndStep(e.result.Outcome); err != nil {
		// TODO logging
		e.result.Outcome = dossiers.ResultFailure
	}
}

func (e *stepExecutor) endTask(ctx context.Context, task *Task) {
	if e.result.Outcome == dossiers.ResultFailure {
		base := e.stepRun.Base()
		evalSupplier := &evaluationSupplier{dossier: e.dossier}
		if continueOnError, err := base.ContinueOnError.Evaluate("job.step.continue-on-error", evalSupplier); err != nil {
			//logger.Infof("Failed but continue next step")
			//return err
			e.result.Conclusion = dossiers.ResultFailure
		} else if continueOnError {
			//logger.Infof("Failed but continue next step")
			e.result.Conclusion = dossiers.ResultSuccess
		} else {
			e.result.Conclusion = dossiers.ResultFailure
		}
	}

	if e.parent == nil { // root step
		e.reporter.EndStep(e.StepId(), e.result.Outcome)
	}
}

func (e *stepExecutor) ComposeEnv() map[string]string {
	// clone dossier.Env to avoid modifying
	m := maps.Clone(e.dossier.Env)

	// NOTE:
	// * INPUT_* env will be set in the step task
	// * Other default envs are set when sandbox is created

	// set STATE_* env
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
	for k, v := range e.state {
		k = "STATE_" + k
		m[k] = v
	}

	// set GITHUB_ACTION_* env
	gh := e.dossier.Github
	m["GITHUB_ACTION"] = e.stepRun.StepId()
	m["GITHUB_ACTION_REF"] = gh.ActionRef
	m["GITHUB_ACTION_REPOSITORY"] = gh.ActionRepository

	// set file commands env
	maps.Copy(m, e.extraEnv)

	m["PATH"] = e.job.ComposePath()

	return m
}

func (e *stepExecutor) SetResult(outcome dossiers.Result) {
	e.result.Outcome = outcome
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *stepExecutor) SetEnv(env map[string]string, inExprValue bool) error {
	for k := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			delete(env, k)
		}
	}
	// env not added to the expr evaluation context
	if !inExprValue {
		maps.Copy(e.extraEnv, env)
	} else {
		// add env to both step-level and job-level context
		maps.Copy(e.dossier.Env, env)
		if jh, ok := e.job.(JobCommandHandler); ok {
			return jh.SetEnv(env)
		}
	}
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (e *stepExecutor) CreateStepSummary() error {
	panic("implement me")
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) error {
	if e.parent != nil {
		if root, ok := e.RootExecutor().(StepCommandHandler); ok {
			return root.SaveState(state)
		}
	}
	maps.Copy(e.state, state)
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *stepExecutor) SetOutput(output map[string]string) error {
	maps.Copy(e.result.Outputs, output)
	return nil
}
