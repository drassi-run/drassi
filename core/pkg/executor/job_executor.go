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
	"maps"
	"path"
	"slices"
	"strings"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"drassi.run/core/util/tar"
	"go.uber.org/dig"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobExecutor interface {
	JobSpec() *JobSpec

	Initialize(ctx context.Context) (*records.Job, error)
	RunJob(ctx context.Context) (*records.Job, error)
	Finalize(ctx context.Context) (*records.Job, error)

	Sandbox() sandboxer.Sandbox
	ExprEnv() expression.Env

	Github() *records.Github
	Status() records.Result // inherit libraries.StatusProvider
	SetStatus(status records.Result)
	AddPath(paths []string)
	SystemPaths() []string
	Env() map[string]string
	SystemEnv() map[string]string
	SetEnv(env map[string]string)
}

type JobTask struct {
	Run      JobRun
	Stage    Stage
	Executor JobExecutor
}

func (t *JobTask) JobId() string {
	return t.JobSpec().Id
}

func (t *JobTask) JobSpec() *JobSpec {
	return t.Executor.JobSpec()
}

type JobRun func(context.Context) (*records.Job, error)

type jobExecutor struct {
	spec  *JobSpec
	scope *dig.Scope

	// records
	github  *records.Github
	jobInfo *records.JobInfo
	job     *records.Job
	steps   map[string]*records.Step
	env     map[string]string
	paths   []string

	postStart Hook[JobExecutor]
	preStop   Hook[JobExecutor]
	decorator StepRunDecorator
	exprEnv   expression.Env
	sandbox   sandboxer.Sandbox

	children  map[string]StepExecutor
	stageRuns map[Stage]func(context.Context) error
}

func (e *jobExecutor) JobSpec() *JobSpec {
	return e.spec
}

func (e *jobExecutor) Initialize(ctx context.Context) (job *records.Job, err error) {
	// e.job is late initialize
	defer func() { job = e.job }()
	s := scribe.FromContext(ctx)

	// do job initialization
	if ex := e.initializeJob(s); ex != nil {
		err = fmt.Errorf("initialize job: %w", ex)
		return
	}

	if ex := e.initializeSandbox(ctx, s); ex != nil {
		err = fmt.Errorf("initialize sandbox: %w", ex)
		return
	}

	if ex := e.initializeScope(); ex != nil {
		err = fmt.Errorf("initialize scope: %w", ex)
		return
	}

	if ex := e.initializeSteps(ctx, s); ex != nil {
		err = fmt.Errorf("initialize steps: %w", ex)
		return
	}

	if hook := e.postStart; hook != nil {
		if ex := hook.Hook(ctx, e); ex != nil {
			err = fmt.Errorf("postStart hook: %w", ex)
		}
	}

	s.Writef("Setup job done")
	return
}

func (e *jobExecutor) RunJob(ctx context.Context) (*records.Job, error) {
	errs := make([]error, 3)
	for i, stage := range []Stage{StagePre, StageMain, StagePost} {
		if run := e.stageRuns[stage]; run != nil {
			errs[i] = run(ctx)
		}
	}

	return e.job, errors.Join(errs...)
}

func (e *jobExecutor) Finalize(ctx context.Context) (job *records.Job, err error) {
	errs := make([]error, 0, 3)
	defer func() {
		job = e.job
		err = errors.Join(errs...)
	}()

	s := scribe.FromContext(ctx)
	if e.job.Result == records.ResultSuccess {
		s.Debugf("Evaluating job 'outputs'")
		if ex := evaluator.Evaluate(e.exprEnv, e.spec.Outputs, &e.job.Outputs); ex != nil {
			e.SetStatus(records.ResultFailure)
			s.Errorf("Evaluate job 'outputs' error: %v", ex)
			ex = fmt.Errorf("evaluate job 'output': %w", ex)
			errs = append(errs, ex)
		}
	}

	if hook := e.preStop; hook != nil {
		if ex := hook.Hook(ctx, e); ex != nil {
			ex = fmt.Errorf("preStop hook: %w", ex)
			s.Errorf("preStop hook error: %v", ex)
			errs = append(errs, ex)
		}
	}

	if e.sandbox == nil {
		return
	}

	s.Writef("Terminating sandbox...")
	if ex := e.sandbox.Terminate(ctx); ex != nil {
		s.Errorf("Sandbox terminate error: %v", err)
		ex = fmt.Errorf("terminate sandbox: %w", err)
		errs = append(errs, ex)
	} else {
		s.Writef("Sandbox terminated")
	}

	s.Writef("Teardown job done")
	return
}

type paramHooks struct {
	dig.In

	PostStart Hook[JobExecutor] `name:"post-start" optional:"true"` // see [wire.PostStart]
	PreStop   Hook[JobExecutor] `name:"pre-stop" optional:"true"`   // see [wire.PreStop]
}

func (e *jobExecutor) populateHooks(p paramHooks) {
	e.postStart = p.PostStart
	e.preStop = p.PreStop
}

func (e *jobExecutor) initializeJob(s *scribe.Scribe) error {
	// inject dependencies
	// workaround because dig not support pass dig.Name() into Invoke func
	if err := e.scope.Invoke(e.populateHooks); err != nil {
		return fmt.Errorf("populate 'hooks': %w", err)
	}
	if err := xdig.Populate(e.scope, &e.decorator); err != nil {
		return fmt.Errorf("populate 'decorator': %w", err)
	}
	if err := xdig.Populate(e.scope, &e.exprEnv); err != nil {
		return fmt.Errorf("populate 'exprEnv': %w", err)
	}
	if err := xdig.Populate(e.scope, &e.env); err != nil {
		return fmt.Errorf("populate 'env': %w", err)
	}
	if err := xdig.Populate(e.scope, &e.github); err != nil {
		return fmt.Errorf("populate 'github': %w", err)
	}

	// sanitize GitHub
	e.github = new(*e.github) // clone
	e.github.Job = e.spec.Id
	e.github.Action = ""
	e.github.ActionPath = ""
	e.github.ActionRef = ""
	e.github.ActionRepository = ""
	e.github.ActionStatus = ""
	e.jobInfo = new(records.JobInfo)
	e.job = new(records.Job)
	e.steps = make(map[string]*records.Step, len(e.spec.Steps))
	e.SetStatus(records.ResultSuccess)

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", e.github),
		expression.WithVariable("job", e.jobInfo),
		expression.WithVariable("steps", e.steps),
		expression.WithVariable("env", e.env),
		expression.WithLibrary(libraries.StatusLib(e)),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
		return fmt.Errorf("create child expression.Env: %w", err)
	} else {
		e.exprEnv = exprEnv
	}

	// Evaluate expressions
	s.Debugf("Evaluating job 'env'")
	env := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Env, &env); err != nil {
		s.Errorf("Evaluate job 'env' error: %v", err)
		return fmt.Errorf("evaluate job 'env': %w", err)
	} else {
		e.SetEnv(env)
	}

	s.Debugf("Evaluating job 'defaults'")
	var defaults workflows.Defaults
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Defaults, &defaults); err != nil {
		s.Errorf("Evaluate job 'defaults' error: %v", err)
		return fmt.Errorf("evaluate job 'defaults': %w", err)
	} else if err = xdig.Supply(e.scope, defaults); err != nil {
		return fmt.Errorf("supply 'defaults': %w", err)
	}

	return nil
}

func (e *jobExecutor) initializeSandbox(ctx context.Context, s *scribe.Scribe) error {
	var runtime sandboxer.Engine
	if err := xdig.Populate(e.scope, &runtime); err != nil {
		return err
	}

	req := &sandboxer.LaunchRequest{
		Uid:    e.spec.Uid,
		Github: e.github,
	}

	s.Debugf("Evaluating 'container'")
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Container, &req.JobContainer); err != nil {
		s.Errorf("Evaluate 'container' error: %v", err)
		return fmt.Errorf("evaluate 'container': %w", err)
	}

	s.Debugf("Evaluating 'services' containers")
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Services, &req.ServiceContainers); err != nil {
		s.Errorf("Evaluate 'services' containers error: %v", err)
		return fmt.Errorf("evaluate 'services': %w", err)
	}

	s.Writef("Launching sandbox...")
	if resp, err := runtime.Launch(ctx, req); err != nil {
		s.Errorf("Sandbox launch failed: %v", err)
		return fmt.Errorf("launch sandbox: %w", err)
	} else {
		s.Writef("Sandbox is ready")
		e.sandbox = resp.Sandbox

		// set records values
		s.Debugf("Update context data")
		e.jobInfo.Container = resp.JobContainer
		e.jobInfo.Services = resp.ServiceContainers

		layout := e.sandbox.Layout()
		e.github.Workspace = layout.Workspace

		if err = xdig.Supply(e.scope, resp.ContainerEngine, dig.Export(true)); err != nil {
			return fmt.Errorf("supply 'container.Engine': %w", err)
		}
		if err = xdig.Supply(e.scope, resp.Sandbox, dig.Export(true)); err != nil {
			return fmt.Errorf("supply 'sandbox': %w", err)
		}
	}

	if location, err := e.setupEventFile(ctx); err != nil {
		return err
	} else {
		e.github.EventPath = location
	}

	// register SandboxLib (e.g. hashFiles func) to expression.Env
	var cp xcontext.Provider
	if err := xdig.Populate(e.scope, &cp); err != nil {
		return err
	}
	opt := expression.WithLibrary(libraries.SandboxLib(cp, e.sandbox))
	if exprEnv, err := e.exprEnv.New(opt); err != nil {
		return fmt.Errorf("create child expression.Env: %w", err)
	} else {
		e.exprEnv = exprEnv
	}

	return nil
}

func (e *jobExecutor) initializeScope() error {
	// Provide scope values
	if err := e.scope.Invoke(e.configureRunner); err != nil {
		return fmt.Errorf("configure records.Runner: %w", err)
	}
	if err := xdig.Supply(e.scope, e.exprEnv); err != nil {
		return fmt.Errorf("supply 'exprEnv': %w", err)
	}
	if err := xdig.Supply(e.scope, e.github); err != nil {
		return fmt.Errorf("supply 'github': %w", err)
	}
	if err := xdig.Supply(e.scope, e.jobInfo, dig.Export(true)); err != nil {
		return fmt.Errorf("supply 'jobInfo': %w", err)
	}
	if err := xdig.Supply(e.scope, e.env); err != nil {
		return fmt.Errorf("supply 'env': %w", err)
	}

	return nil
}

func (e *jobExecutor) configureRunner(runner *records.Runner) {
	layout := e.sandbox.Layout()
	runner.Workspace = layout.Workspace
	runner.ToolCache = layout.Tools
	runner.Temp = layout.Temp
}

func (e *jobExecutor) initializeSteps(ctx context.Context, s *scribe.Scribe) error {
	e.children = make(map[string]StepExecutor, len(e.spec.Steps))

	// TODO: concurrent version of Initialize is temporary disable because of concurrent map writes in scope
	//g, ctx := errgroup.WithContext(ctx)
	//for _, step := range e.jobRun.Steps {
	//	exec := e.NewStepExecutor(step)
	//	s := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
	//	g.Go(func() error {
	//		return exec.Initialize(s)
	//	})
	//}
	//return g.Wait()

	s.Writef("Initialize steps...")
	for _, step := range e.spec.Steps {
		scope := e.scope.Scope(fmt.Sprintf("step(%s)", step.Id))
		if exec, err := step.CreateExecutor(ctx, scope, e, nil); err != nil {
			return fmt.Errorf("create StepExecutor for %q: %w", step.Id, err)
		} else {
			e.children[step.Id] = exec
		}
	}
	s.Writef("Steps initialized")

	s.Writef("Planing steps execution...")
	e.stageRuns = make(map[Stage]func(context.Context) error, 3)
	e.stageRuns[StagePre] = e.planStage(StagePre)
	e.stageRuns[StageMain] = e.planStage(StageMain)
	e.stageRuns[StagePost] = e.planStage(StagePost)
	s.Writef("Steps execution planed")
	return nil
}

func (e *jobExecutor) planStage(stage Stage) func(context.Context) error {
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
			task.Run = e.telemetry(id, stage, task.Run)
		}
		tasks[i] = task
	}

	// all runs are nil
	if !slices.ContainsFunc(tasks, func(t *StepTask) bool { return t != nil }) {
		return nil
	}

	// 2. Execute the plan
	return func(ctx context.Context) error {
		errs := make([]error, 0)
		for i, task := range tasks {
			if task == nil {
				continue
			}

			id := ids[i]
			res, err := task.Run(ctx)
			// Only set `steps` records in `main` stage & `id` is user specified
			if stage == StageMain && !strings.HasPrefix(id, "__") {
				e.steps[id] = res
			}
			if res != nil {
				if con := res.Conclusion; weight(con) > weight(e.job.Result) {
					e.SetStatus(con)
				}
			}
			if err != nil {
				errs = append(errs, fmt.Errorf("run step %q (%s): %w", id, stage, err))
			}
		}
		return errors.Join(errs...)
	}
}

func (e *jobExecutor) telemetry(stepId string, stage Stage, run StepRun) StepRun {
	return func(ctx context.Context) (_ *records.Step, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s/%s)", stage, stepId),
			xotel.Step(string(stage)+"/"+stepId),
		)
		defer done(&err)

		return run(ctx)
	}
}

func (e *jobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *jobExecutor) ExprEnv() expression.Env {
	return e.exprEnv
}

func (e *jobExecutor) Github() *records.Github {
	return e.github
}

// Status return current job's Result
// its implement [libraries.StatusProvider]
func (e *jobExecutor) Status() records.Result {
	if e.job != nil {
		return e.job.Result
	}
	return records.ResultSuccess
}

func (e *jobExecutor) SetStatus(status records.Result) {
	e.job.Result = status
	e.jobInfo.Status = status
}

// AddPath prepending a directory to the system PATH variable (and remove duplicates).
// It automatically makes it available to all subsequent actions in the current job.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path
func (e *jobExecutor) AddPath(paths []string) {
	if len(paths) == 0 {
		return
	}

	newPaths := make([]string, 0, len(e.paths))
	set := sets.New[string]()

	for _, p := range slices.Backward(paths) {
		if !set.Has(p) {
			newPaths = append(newPaths, p)
			set.Insert(p)
		}
	}
	for _, p := range e.paths {
		if !set.Has(p) {
			newPaths = append(newPaths, p)
			set.Insert(p)
		}
	}

	e.paths = newPaths
}

func (e *jobExecutor) SystemPaths() []string {
	return e.paths
}

func (e *jobExecutor) Env() map[string]string {
	return e.env
}

func (e *jobExecutor) SystemEnv() map[string]string {
	return nil
}

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *jobExecutor) SetEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

func (e *jobExecutor) setupEventFile(ctx context.Context) (string, error) {
	tmp := e.sandbox.Layout().Temp
	files := map[string]any{"workflow/event.json": e.github.Event}
	r, err := xtar.JsonObjectReader(files, false)
	if err != nil {
		return "", err
	}

	if err = e.sandbox.CopyIn(ctx, r, tmp); err != nil {
		return "", fmt.Errorf("copy event file to sandbox: %w", err)
	}

	location := path.Join(tmp, "workflow", "event.json")
	return location, nil
}
