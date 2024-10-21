package executor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/tar"
	"go.uber.org/dig"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobExecutor interface {
	JobRun() *JobRun
	Context() context.Context
	SetContext(ctx context.Context)

	Initialize(scope *dig.Scope) error
	RunJob() *records.Job
	Finalize() error
	SystemPaths() []string
	SetStatus(status records.Result)

	AddPath(paths []string) error
	SetEnv(env map[string]string) error
}

func JobId(e JobExecutor) string {
	return e.JobRun().Id
}

func JobUid(e JobExecutor) string {
	return e.JobRun().Uid
}

func NewJobExecutor(run *JobRun) JobExecutor {
	return &jobExecutor{
		jobRun: run,
	}
}

type jobExecutor struct {
	jobRun *JobRun

	// records
	github  records.Github
	runner  records.Runner
	jobInfo *records.JobInfo
	job     *records.Job
	steps   map[string]*records.Step
	env     map[string]string
	paths   []string

	ctx           context.Context
	exprEnv       expression.Env
	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor
	supervisor    Supervisor
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

func (e *jobExecutor) RunJob() *records.Job {
	e.SetStatus(records.ResultSuccess)

	if err := e.runStage(StagePre, StepRun.PreTask); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := e.runStage(StageMain, StepRun.MainTask); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := e.runStage(StagePost, StepRun.PostTask); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Outputs, &e.job.Outputs); err != nil {
		e.SetStatus(records.ResultFailure)
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

	return e.sandbox.Terminate(ctx)
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
		e.github.Job = JobId(e)
		e.github.Action = ""
		e.github.ActionPath = ""
		e.github.ActionRef = ""
		e.github.ActionRepository = ""
		e.github.ActionStatus = ""
	}
	e.jobInfo = new(records.JobInfo)
	e.job = new(records.Job)
	e.steps = make(map[string]*records.Step, len(e.jobRun.Steps))

	if err := e.supervisor.BeforeJobRun(e); err != nil {
		return err
	}

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("job", e.jobInfo),
		expression.WithVariable("steps", e.steps),
		expression.WithVariable("env", e.env),
		expression.WithLibrary(libraries.JobLib(e.job)),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
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

	var defaults workflows.Defaults
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Defaults, &defaults); err != nil {
		return err
	} else if err = xdig.Supply(scope, defaults); err != nil {
		return err
	}

	return nil
}

func (e *jobExecutor) initializeSandbox(scope *dig.Scope) error {
	var runtime sandboxer.Engine
	if err := xdig.Populate(scope, &runtime); err != nil {
		return err
	}

	req := &sandboxer.LaunchRequest{
		Uid:    JobUid(e),
		Github: &e.github,
	}

	var jobContainer *workflows.Container
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Container, &jobContainer); err != nil {
		return err
	} else {
		req.JobContainer = jobContainer
	}

	var serviceContainers = make(map[string]*workflows.Container)
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Services, &serviceContainers); err != nil {
		return err
	} else if len(serviceContainers) > 0 {
		req.ServiceContainers = serviceContainers
	}

	if resp, err := runtime.Launch(e.ctx, req); err != nil {
		return err
	} else {
		e.sandbox = resp.Sandbox

		// set records values
		e.jobInfo.Container = resp.JobContainer
		e.jobInfo.Services = resp.ServiceContainers

		layout := e.sandbox.Layout()
		e.github.Workspace = layout.Workspace
		e.runner.Workspace = layout.Workspace
		e.runner.ToolCache = layout.Tools
		e.runner.Temp = layout.Temp

		if err = xdig.Supply(scope, resp.ContainerEngine, dig.Export(true)); err != nil {
			return err
		}
		if err = xdig.Supply(scope, resp.Sandbox, dig.Export(true)); err != nil {
			return err
		}
	}

	if location, err := e.setupEventFile(e.ctx); err != nil {
		return err
	} else {
		e.github.EventPath = location
	}

	// register SandboxLib (e.g. hashFiles func) to expression.Env
	opts := []expression.Option{
		expression.WithLibrary(libraries.SandboxLib(e.supervisor, e.sandbox)),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
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
	if err := xdig.Supply(scope, e.jobInfo, dig.Export(true)); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.env); err != nil {
		return err
	}

	// initialize ConsoleCommand & FileCommand
	// NOTE: some handlers are depended on sandbox
	if err := scope.Invoke(func(p consoleCommandParams) error {
		for _, h := range p.Handlers {
			if err := p.CmdMgr.Register(h); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := scope.Invoke(func(p fileCommandParams) error {
		for _, h := range p.Handlers {
			if err := p.CmdMgr.Register(h); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if e.runner.Debug == "1" {
		if err := scope.Invoke(func(cmdMgr command.ConsoleManager) error {
			cmd := &command.Command{Name: "echo", Value: "ON"}
			return cmdMgr.Process("", cmd)
		}); err != nil {
			return err
		}
	}

	err := scope.Invoke(func(cmdMgr command.FileManager) error {
		cb := func(_ *Task, exec StepExecutor) error {
			ctx := exec.Context()
			suffix := StepUid(exec)
			return cmdMgr.Initialize(ctx, suffix)
		}
		e.supervisor.Register(BeforeRunTaskCallback(cb))

		cb = func(_ *Task, exec StepExecutor) error {
			ctx := exec.Context()
			suffix := StepUid(exec)
			return cmdMgr.Process(ctx, suffix)
		}
		e.supervisor.Register(AfterRunTaskCallback(cb))

		ep := func() map[string]string {
			exec := e.supervisor.CurrentStep()
			if exec == nil {
				return nil
			}

			suffix := StepUid(exec)
			return cmdMgr.Env(suffix)
		}
		e.supervisor.Register(EnvProvider(ep))

		return nil
	})

	runnerEnv := map[string]string{
		"RUNNER_NAME":        e.runner.Name,
		"RUNNER_ARCH":        string(e.runner.Arch),
		"RUNNER_OS":          string(e.runner.Os),
		"RUNNER_ENVIRONMENT": e.runner.Environment,
		"RUNNER_TEMP":        e.runner.Temp,
		"RUNNER_TOOL_CACHE":  e.runner.ToolCache,
		"RUNNER_WORKSPACE":   e.runner.Workspace,
	}
	if e.runner.Debug == "1" {
		runnerEnv["RUNNER_DEBUG"] = "1"
	}
	e.supervisor.Register(Env(runnerEnv))

	return err
}

type consoleCommandParams struct {
	dig.In

	CmdMgr   command.ConsoleManager
	Handlers []*command.ConsoleHandler `group:"console-handlers"`
}

type fileCommandParams struct {
	dig.In

	CmdMgr   command.FileManager
	Handlers []*command.FileHandler `group:"file-handlers"`
}

func (e *jobExecutor) initializeSteps(scope *dig.Scope) error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.jobRun.Steps))

	// TODO: concurrent version of Initialize is temporary disable because of concurrent map writes in scope
	//g, ctx := errgroup.WithContext(e.ctx)
	//for _, step := range e.jobRun.Steps {
	//	exec := e.NewStepExecutor(step)
	//	s := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
	//	g.Go(func() error {
	//		exec.SetContext(ctx)
	//		return exec.Initialize(s)
	//	})
	//}
	//return g.Wait()

	for _, step := range e.jobRun.Steps {
		exec := e.NewStepExecutor(step)
		s := scope.Scope(fmt.Sprintf("step(%s)", StepId(exec)))
		exec.SetContext(e.ctx)
		if err := exec.Initialize(s); err != nil {
			return err
		}
	}
	return nil
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
		// Only set `steps` records in `main` stage & `id` is user specified
		if stage == StageMain && !strings.HasPrefix(id, "__") {
			e.steps[id] = res
		}
		if res.Conclusion == records.ResultFailure {
			e.job.Result = records.ResultFailure
			return fmt.Errorf(`step %q (%s) failed`, id, stage)
		}
	}

	return nil
}

func (e *jobExecutor) SystemPaths() []string {
	return slices.Clone(e.paths)
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
func (e *jobExecutor) AddPath(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	newPaths := make([]string, 0, len(e.paths))
	set := sets.New[string]()

	for _, path := range slices.Backward(paths) {
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

func (e *jobExecutor) setupEventFile(ctx context.Context) (string, error) {
	files := map[string]any{"": e.github.Event}
	r, err := xtar.JsonObjectReader(files, false)
	if err != nil {
		return "", err
	}

	location := filepath.Join(e.runner.Temp, "workflow", "event.json")
	if err = e.sandbox.CopyIn(ctx, r, location); err != nil {
		return "", err
	}

	return location, nil
}
