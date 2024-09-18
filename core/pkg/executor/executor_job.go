package executor

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/util/dig"

	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobExecutor interface {
	JobId() string
	Context() context.Context
	SetContext(ctx context.Context)

	Initialize(scope *dig.Scope) error
	RunJob() *dossiers.Job
	Finalize() error
	ComposeEnv(m map[string]string)
	SetStatus(status dossiers.Result)

	AddPath(paths []string) error
	SetEnv(env map[string]string) error
}

func NewJobExecutor(run *JobRun) JobExecutor {
	return &jobExecutor{
		jobRun: run,
	}
}

type jobExecutor struct {
	jobRun *JobRun

	// records
	github  dossiers.Github
	runner  dossiers.Runner
	jobInfo *dossiers.JobInfo
	job     *dossiers.Job
	steps   map[string]*dossiers.Step
	env     map[string]string
	paths   []string

	ctx           context.Context
	exprEnv       *expression.Env
	runtime       sandboxer.SandboxRuntime
	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor
	supervisor    Supervisor
}

func (e *jobExecutor) JobId() string {
	return e.jobRun.Id
}

func (e *jobExecutor) JobRun() *JobRun {
	return e.jobRun
}

func (e *jobExecutor) NewStepExecutor(step StepRun) StepExecutor {
	exec := newStepExecutor(e, nil, step)
	e.stepExecutors[step.StepId()] = exec
	return exec
}

func (e *jobExecutor) StepExecutor(id string) StepExecutor {
	return e.stepExecutors[id]
}

func (e *jobExecutor) Context() context.Context {
	return e.ctx
}

func (e *jobExecutor) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *jobExecutor) Initialize(scope *dig.Scope) error {
	if err := e.initializeJob(scope); err != nil {
		return err
	}

	if err := e.initializeSandbox(scope); err != nil {
		return err
	}

	if err := e.initializeScope(scope); err != nil {
		return err
	}

	return e.initializeSteps(scope)
}

func (e *jobExecutor) RunJob() *dossiers.Job {
	e.SetStatus(dossiers.ResultSuccess)
	states := map[Stage]func(StepRun) *Task{
		StagePre:  StepRun.PreTask,
		StageMain: StepRun.MainTask,
		StagePost: StepRun.PostTask,
	}
	for state, fn := range states {
		if err := e.runStage(state, fn); err != nil {
			e.SetStatus(dossiers.ResultFailure)
			//log err
		}
	}
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Outputs, &e.job.Outputs); err != nil {
		e.SetStatus(dossiers.ResultFailure)
	}
	return e.job
}

func (e *jobExecutor) Finalize() (err error) {
	defer func() {
		errs := make([]error, 0)
		if err != nil {
			errs = append(errs, err)
		}

		if err := e.supervisor.AfterJobRun(e, e.job); err != nil {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			err = errors.Join(errs...)
		}
	}()

	if e.sandbox == nil {
		return
	}

	// if ctx is done, a new one is created w/ timeout 5s to clean up resources
	ctx := e.ctx
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	req := sandboxer.TerminateSandboxRequest{
		Sandbox: e.sandbox,
	}
	_, err = e.runtime.TerminateSandbox(ctx, req)
	return
}

func (e *jobExecutor) initializeJob(scope *dig.Scope) error {
	// inject dependencies
	if err := xdig.Populate(scope, &e.supervisor); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.env); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.runner); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.github); err != nil {
		return err
	} else {
		// sanitize GitHub
		e.github.Job = e.JobId()
		e.github.Action = ""
		e.github.ActionPath = ""
		e.github.ActionRef = ""
		e.github.ActionRepository = ""
		e.github.ActionStatus = ""
	}
	e.jobInfo = new(dossiers.JobInfo)
	e.job = new(dossiers.Job)
	e.steps = make(map[string]*dossiers.Step, len(e.jobRun.Steps))

	if err := e.supervisor.BeforeJobRun(e); err != nil {
		return err
	}

	// setup expression.Env
	opts := []expression.EnvOption{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("job", e.jobInfo),
		expression.WithVariable("steps", e.steps),
		expression.WithVariable("env", e.env),
		expression.WithLibrary(libraries.JobLib(e.jobInfo)),
	}
	if exprEnv, err := expression.NewEnv(e.exprEnv, opts...); err != nil {
		return err
	} else {
		e.exprEnv = exprEnv
	}

	// Evaluate expressions
	env := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Env, &env); err != nil {
		return err
	} else if err = e.SetEnv(env); err != nil {
		return err
	}

	var defaults *workflows.Defaults
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Defaults, &defaults); err != nil {
		return err
	} else if defaults != nil {
		if err = xdig.Supply(scope, defaults); err != nil {
			return err
		}
	}

	return nil
}

func (e *jobExecutor) initializeSandbox(scope *dig.Scope) error {
	req := sandboxer.LaunchSandboxRequest{JobId: e.jobRun.Id}

	var jobContainer *workflows.Container
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Container, &jobContainer); err != nil {
		return err
	} else if jobContainer != nil {
		if con, err := e.toContainerConfig(e.ctx, jobContainer); err != nil {
			return err
		} else {
			req.JobContainer = con
		}
	}

	var serviceContainers = make(map[string]*workflows.Container)
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Services, &serviceContainers); err != nil {
		return err
	} else if len(serviceContainers) > 0 {
		services := make(map[string]*container.ContainerConfig, len(serviceContainers))
		for name, srv := range serviceContainers {
			if con, err := e.toContainerConfig(e.ctx, srv); err != nil {
				return err
			} else {
				services[name] = con
			}
		}
		req.ServiceContainers = services
	}

	if resp, err := e.runtime.LaunchSandbox(e.ctx, req); err != nil {
		return err
	} else {
		e.sandbox = resp.Sandbox
		e.jobInfo.Container = resp.Container
		e.jobInfo.Services = resp.Services
	}

	// register SandboxLib (e.g hashFiles func) to expression.Env
	opts := []expression.EnvOption{
		expression.WithLibrary(libraries.SandboxLib(e.supervisor, e.sandbox)),
	}
	if exprEnv, err := expression.NewEnv(e.exprEnv, opts...); err != nil {
		return err
	} else {
		e.exprEnv = exprEnv
	}

	return nil
}

func (e *jobExecutor) initializeScope(scope *dig.Scope) error {
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
	if err := xdig.Supply(scope, e.sandbox, dig.Export(true)); err != nil {
		return err
	}

	// initialize ConsoleCommand & FileCommand
	// NOTE: some handlers are depended on sandbox
	if err := scope.Invoke(func(p consoleCommandParams) error {
		for _, h := range p.handlers {
			if err := p.cmdMgr.Register(h); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := scope.Invoke(func(p fileCommandParams) error {
		for _, h := range p.handlers {
			if err := p.cmdMgr.Register(h); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	err := scope.Invoke(func(cmdMgr command.FileManager) error {
		cb := func(_ *Task, exec StepExecutor) error {
			ctx := exec.Context()
			suffix := exec.StepId()
			return cmdMgr.Initialize(ctx, suffix)
		}
		e.supervisor.Register(BeforeRunTaskCallback(cb))

		cb = func(_ *Task, exec StepExecutor) error {
			ctx := exec.Context()
			suffix := exec.StepId()
			return cmdMgr.Process(ctx, suffix)
		}
		e.supervisor.Register(AfterRunTaskCallback(cb))

		ep := func() map[string]string {
			exec := e.supervisor.CurrentStep()
			if exec == nil {
				return nil
			}

			suffix := exec.StepId()
			return cmdMgr.Env(suffix)
		}
		e.supervisor.Register(EnvProvider(ep))

		return nil
	})

	return err
}

type consoleCommandParams struct {
	dig.In

	cmdMgr   command.ConsoleManager
	handlers []*command.ConsoleHandler `group:"console-handlers"`
}

type fileCommandParams struct {
	dig.In

	cmdMgr   command.FileManager
	handlers []*command.FileHandler `group:"file-handlers"`
}

func (e *jobExecutor) initializeSteps(scope *dig.Scope) error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.jobRun.Steps))
	g, ctx := errgroup.WithContext(e.ctx)
	for _, step := range e.jobRun.Steps {
		exec := e.NewStepExecutor(step)
		s := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
		g.Go(func() error {
			exec.SetContext(ctx)
			return exec.Initialize(s)
		})
	}
	return g.Wait()
}

func (e *jobExecutor) runStage(stage Stage, fn func(StepRun) *Task) error {
	ids := make([]string, len(e.jobRun.Steps))
	for i, step := range e.jobRun.Steps {
		ids[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(ids) // in place reverse
	}
	for _, id := range ids {
		exec := e.StepExecutor(id)
		exec.SetContext(e.ctx)
		res := exec.RunStep(fn)
		if res == nil {
			continue
		}
		e.steps[id] = res
		if res.Conclusion == dossiers.ResultFailure {
			return fmt.Errorf(`step %q (%s) failed`, id, stage)
		}
	}

	return nil
}

func (e *jobExecutor) ComposeEnv(m map[string]string) {
	runnerEnv := map[string]string{
		"RUNNER_NAME":        e.runner.Name,
		"RUNNER_ARCH":        string(e.runner.Arch),
		"RUNNER_OS":          string(e.runner.Os),
		"RUNNER_ENVIRONMENT": e.runner.Environment,
	}
	if e.runner.Debug == "1" {
		m["RUNNER_DEBUG"] = "1"
	}
	maps.Copy(m, runnerEnv)

	supEnv := e.supervisor.ProvideEnv()
	maps.Copy(m, supEnv)

	m["PATH"] = strings.Join(e.paths, ":")
}

func (e *jobExecutor) SetStatus(status dossiers.Result) {
	e.job.Result = status
	e.jobInfo.Status = status
}

// AddPath prepending a directory to the system PATH variable (and remove duplicates).
// It automatically makes it available to all subsequent actions in the current job.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path
func (e *jobExecutor) AddPath(paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	slices.Reverse(paths)

	newPaths := make([]string, 0, len(paths))
	set := sets.New(paths[0])

	for _, path := range paths[1:] {
		if !set.Has(path) {
			newPaths = append(newPaths, path)
			set.Insert(path)
		}
	}
	for _, path := range e.paths {
		if !set.Has(path) {
			newPaths = append(newPaths, path)
			set.Insert(path)
		}
	}

	e.paths = newPaths
	return nil
}

var setEnvBlockList = sets.New("NODE_OPTIONS")

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *jobExecutor) SetEnv(env map[string]string) error {
	for k, v := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			continue
		}
		e.env[k] = v
	}
	return nil
}
