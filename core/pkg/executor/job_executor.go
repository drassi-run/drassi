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
	"time"

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
	CreateTask() *JobTask

	Sandbox() sandboxer.Sandbox
	ExprEnv() expression.Env

	Forge() *records.Forge
	Path() []string
	AddPath(paths []string)
	Env() map[string]string
	SetEnv(env map[string]string)
}

type JobTask struct {
	Run      JobRun
	Executor JobExecutor
}

func (t *JobTask) JobId() string {
	return t.JobSpec().Id
}

func (t *JobTask) JobSpec() *JobSpec {
	return t.Executor.JobSpec()
}

type JobRun func(context.Context) (*records.JobResult, error)

type jobExecutor struct {
	spec  *JobSpec
	scope *dig.Scope

	// records
	forge   *records.Forge
	jobInfo *records.JobInfo
	job     *records.JobResult
	steps   map[string]*records.StepResult
	env     map[string]string
	paths   []string

	postStart Hook[JobExecutor]
	preStop   Hook[JobExecutor]
	decorator StepRunDecorator
	exprEnv   expression.Env
	sandbox   sandboxer.Sandbox

	children  map[string]StepExecutor
	stageRuns map[Stage][]*StepTask
}

func (e *jobExecutor) JobSpec() *JobSpec {
	return e.spec
}

func (e *jobExecutor) CreateTask() *JobTask {
	run := func(ctx context.Context) (*records.JobResult, error) {
		e.job = &records.JobResult{
			Result:  records.ResultSuccess,
			Outputs: make(map[string]string),
		}

		errs := e.run(ctx)
		err := errors.Join(errs...)
		return e.job, err
	}

	return &JobTask{
		Run:      run,
		Executor: e,
	}
}

func (e *jobExecutor) run(ctx context.Context) (errs []error) {
	errs = make([]error, 0, 5)

	err := e.runInitialize(ctx)
	if err != nil {
		errs = append(errs, err)
	}

	defer func() {
		err = e.runFinalize(ctx)
		if err != nil {
			errs = append(errs, err)
		}
	}()

	for _, stage := range []Stage{StagePre, StageMain, StagePost} {
		for _, task := range e.stageRuns[stage] {
			if err = e.runStep(ctx, task); err != nil {
				errs = append(errs, err)
			}
		}
	}

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

func (e *jobExecutor) injectDeps(ctx context.Context) error {
	s := scribe.FromContext(ctx)

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
	if err := xdig.Populate(e.scope, &e.forge); err != nil {
		return fmt.Errorf("populate 'forge': %w", err)
	}

	// sanitize Forge
	e.forge = new(*e.forge) // clone
	e.forge.Job = e.spec.Id
	e.forge.Action = ""
	e.forge.ActionPath = ""
	e.forge.ActionRef = ""
	e.forge.ActionRepository = ""
	e.forge.ActionStatus = ""
	e.jobInfo = new(records.JobInfo)
	e.steps = make(map[string]*records.StepResult, len(e.spec.Steps))

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", e.forge),
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

func (e *jobExecutor) runInitialize(ctx context.Context) error {
	task := &StepTask{
		Run:     runStepE(runActionE(e.initialize)),
		Stage:   StagePre,
		Kind:    workflows.StepKindJob,
		jobSpec: e.spec,
	}
	task.Run = e.decorator.DecorateStepRun(task)
	task.Run = e.telemetry(StagePre, "__init", task.Run)

	res, err := task.Run(ctx)
	if res != nil {
		if con := res.Conclusion; level(e.job.Result) < level(con) {
			e.job.Result = con
		}
	}
	if err != nil {
		err = fmt.Errorf("initialize job: %w", err)
	}
	return err
}

func (e *jobExecutor) initialize(ctx context.Context) error {
	s := scribe.FromContext(ctx)

	if err := e.initializeSandbox(ctx, s); err != nil {
		return fmt.Errorf("initialize sandbox: %w", err)
	}

	if err := e.initializeScope(); err != nil {
		return fmt.Errorf("initialize scope: %w", err)
	}

	if err := e.initializeSteps(ctx, s); err != nil {
		return fmt.Errorf("initialize steps: %w", err)
	}

	if hook := e.postStart; hook != nil {
		if err := hook.Hook(ctx, e); err != nil {
			return fmt.Errorf("postStart hook: %w", err)
		}
	}

	s.Writef("Setup job done")
	return nil
}

func (e *jobExecutor) initializeSandbox(ctx context.Context, s *scribe.Scribe) error {
	var runtime sandboxer.Engine
	if err := xdig.Populate(e.scope, &runtime); err != nil {
		return err
	}

	req := &sandboxer.LaunchRequest{
		Uid:   e.spec.Uid,
		Forge: e.forge,
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
		e.forge.Workspace = layout.Workspace

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
		e.forge.EventPath = location
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
	if err := xdig.Supply(e.scope, e.forge); err != nil {
		return fmt.Errorf("supply 'forge': %w", err)
	}
	if err := xdig.Supply(e.scope, e.jobInfo, dig.Export(true)); err != nil {
		return fmt.Errorf("supply 'jobInfo': %w", err)
	}
	if err := xdig.Supply(e.scope, e.env); err != nil {
		return fmt.Errorf("supply 'env': %w", err)
	}

	return nil
}

func (e *jobExecutor) configureRunner(runner *records.RunnerInfo) {
	layout := e.sandbox.Layout()
	runner.Workspace = layout.Workspace
	runner.ToolCache = layout.Tools
	runner.Temp = layout.Temp
}

func (e *jobExecutor) initializeSteps(ctx context.Context, s *scribe.Scribe) error {
	s.Writef("Initialize steps...")
	e.children = make(map[string]StepExecutor, len(e.spec.Steps))
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
	e.stageRuns = make(map[Stage][]*StepTask, 3)
	for _, stage := range []Stage{StagePre, StageMain, StagePost} {
		e.planStage(s, stage)
	}
	s.Writef("Steps execution planed")
	return nil
}

func (e *jobExecutor) planStage(s *scribe.Scribe, stage Stage) {
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
		s.Debugf("Queue step %s/%s", stage, id)
		tasks = append(tasks, task)
	}
	e.stageRuns[stage] = tasks
}

func (e *jobExecutor) runStep(ctx context.Context, task *StepTask) error {
	stage := task.Stage
	stepId := task.StepId()
	e.jobInfo.Status = e.job.Result

	res, err := task.Run(ctx)

	if res != nil {
		if con := res.Conclusion; level(e.job.Result) < level(con) {
			e.job.Result = con
		}
		// Only set `steps` records in `main` stage & `id` is user specified
		if stage == StageMain && !strings.HasPrefix(stepId, "__") {
			e.steps[stepId] = res
		}
	}

	if err != nil {
		switch e.job.Result {
		case records.ResultFailure:
			err = fmt.Errorf("step %s/%s fail: %w", stage, stepId, err)
		case records.ResultCancelled:
			err = fmt.Errorf("step %s/%s canceled: %w", stage, stepId, err)
		}
	}
	return err
}

func (e *jobExecutor) runFinalize(ctx context.Context) error {
	run := func(ctx context.Context) error {
		errs := e.finalize(ctx)
		return errors.Join(errs...)
	}
	task := &StepTask{
		Run:     runStepE(runActionE(run)),
		Stage:   StagePost,
		Kind:    workflows.StepKindJob,
		jobSpec: e.spec,
	}
	task.Run = e.decorator.DecorateStepRun(task)
	task.Run = e.telemetry(StagePost, "__complete", task.Run)

	// create new ctx w/ timeout 30s to clean up resources
	ctx = context.WithoutCancel(ctx)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	res, err := task.Run(ctx)
	if res != nil {
		if con := res.Conclusion; level(e.job.Result) < level(con) {
			e.job.Result = con
		}
	}
	if err != nil {
		err = fmt.Errorf("complete job: %w", err)
	}
	return err
}

func (e *jobExecutor) finalize(ctx context.Context) (errs []error) {
	s := scribe.FromContext(ctx)

	// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Worker/JobExtension.cs#L751
	// NOTE: any job.Outputs contains secret will be removed later by [wire_secret.maskJobOutputs]
	s.Debugf("Evaluating job 'outputs'")
	if err := evaluator.Evaluate(e.exprEnv, e.spec.Outputs, &e.job.Outputs); err != nil {
		s.Errorf("Evaluate job 'outputs' error: %v", err)
		err = fmt.Errorf("evaluate job 'output': %w", err)
		errs = append(errs, err)
	}

	// TODO: Evaluate environment data

	if hook := e.preStop; hook != nil {
		if err := hook.Hook(ctx, e); err != nil {
			s.Errorf("preStop hook error: %v", err)
			err = fmt.Errorf("preStop hook: %w", err)
			errs = append(errs, err)
		}
	}

	if e.sandbox == nil {
		return
	}

	s.Writef("Terminating sandbox...")
	if err := e.sandbox.Terminate(ctx); err != nil {
		s.Errorf("Sandbox terminate error: %v", err)
		err = fmt.Errorf("terminate sandbox: %w", err)
		errs = append(errs, err)
	} else {
		s.Writef("Sandbox terminated")
	}

	s.Writef("Teardown job done")
	return
}

func (e *jobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *jobExecutor) ExprEnv() expression.Env {
	return e.exprEnv
}

func (e *jobExecutor) Forge() *records.Forge {
	return e.forge
}

// Status return current job's Result
// its implement [libraries.StatusProvider]
func (e *jobExecutor) Status() records.Result {
	if e.job != nil {
		return e.job.Result
	}
	return records.ResultSuccess
}

func (e *jobExecutor) Path() []string {
	return e.paths
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

func (e *jobExecutor) Env() map[string]string {
	return e.env
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
	files := map[string]any{"workflow/event.json": e.forge.Event}
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

func (e *jobExecutor) telemetry(stage Stage, stepId string, run StepRun) StepRun {
	return func(ctx context.Context) (_ *records.StepResult, err error) {
		ctx, done := xotel.SetupTelemetry(ctx,
			fmt.Sprintf("StepRun(%s/%s)", stage, stepId),
			xotel.Step(string(stage)+"/"+stepId),
		)
		defer done(&err)

		return run(ctx)
	}
}
