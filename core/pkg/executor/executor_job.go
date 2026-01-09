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
	"path"
	"slices"
	"strings"
	"time"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/context"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"drassi.run/core/util/tar"
	"github.com/chainguard-dev/clog"
	"go.uber.org/dig"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobExecutor interface {
	JobSpec() *JobSpec

	Initialize(ctx context.Context, scope *dig.Scope) error
	RunJob(ctx context.Context) *records.Job
	Finalize(ctx context.Context) error

	State() *records.Job
	Status() records.Result
	SetStatus(status records.Result)

	AddPath(paths []string)
	SetEnv(env map[string]string)
}

func JobId(e JobExecutor) string {
	return e.JobSpec().Id
}

func JobUid(e JobExecutor) string {
	return e.JobSpec().Uid
}

func NewJobExecutor(spec *JobSpec) JobExecutor {
	return &jobExecutor{jobSpec: spec}
}

type jobExecutor struct {
	jobSpec *JobSpec

	// records
	github  records.Github
	runner  records.Runner
	jobInfo *records.JobInfo
	job     *records.Job
	steps   map[string]*records.Step
	env     map[string]string
	paths   []string

	listener      JobListener
	exprEnv       expression.Env
	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor
}

func (e *jobExecutor) JobSpec() *JobSpec {
	return e.jobSpec
}

func (e *jobExecutor) Initialize(ctx context.Context, scope *dig.Scope) (ex error) {
	jobId := JobId(e)
	ctx, done := xotel.SetupTelemetry(ctx,
		fmt.Sprintf("JobExecutor.Initialize(%s)", jobId),
		xotel.DrassiJob(jobId),
	)
	defer done(&ex)

	l := clog.FromContext(ctx)
	s := scribe.FromContext(ctx)

	// setup listener
	if err := xdig.Populate(scope, &e.listener); err != nil {
		l.Errorf("Failed to initialize job: %v", err)
		return err
	}
	if eh := e.listener.OnInitializeJob(e, scope); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &ex)
		if err != nil {
			return err
		}
	}

	// do job initialization
	if err := e.initializeJob(s, scope); err != nil {
		l.Errorf("Failed to initialize job: %v", err)
		return err
	}

	if err := e.initializeSandbox(ctx, s, scope); err != nil {
		l.Errorf("Failed to initialize sandbox: %v", err)
		return err
	}

	if err := e.initializeScope(scope); err != nil {
		l.Errorf("Failed to initialize scope: %v", err)
		return err
	}

	return e.initializeSteps(ctx, scope)
}

func (e *jobExecutor) RunJob(ctx context.Context) *records.Job {
	jobId := JobId(e)
	ctx, done := xotel.SetupTelemetry(ctx,
		fmt.Sprintf("JobExecutor.RunJob(%s)", jobId),
		xotel.DrassiJob(jobId),
	)
	defer done(nil)

	// setup listener
	if eh := e.listener.OnRunJob(e); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &err)
		if err != nil {
			return nil
		}
	}

	// do job run
	e.SetStatus(records.ResultSuccess)

	if err := e.runStage(ctx, StagePre); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := e.runStage(ctx, StageMain); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := e.runStage(ctx, StagePost); err != nil {
		e.SetStatus(records.ResultFailure)
		//log err
	}
	if err := evaluator.Evaluate(e.exprEnv, e.jobSpec.Outputs, &e.job.Outputs); err != nil {
		e.SetStatus(records.ResultFailure)
	}
	return e.job
}

func (e *jobExecutor) Finalize(ctx context.Context) (ex error) {
	jobId := JobId(e)
	ctx, done := xotel.SetupTelemetry(ctx,
		fmt.Sprintf("JobExecutor.Finalize(%s)", jobId),
		xotel.DrassiJob(jobId),
	)
	defer done(&ex)

	// setup listener
	l := clog.FromContext(ctx)
	if eh := e.listener.OnFinalizeJob(e); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &ex)
		if err != nil {
			l.Errorf("error OnFinalizeJob.Start: %v", err)
			// terminate sandbox even if listener failed
		}
	}

	// do job run
	if e.sandbox == nil {
		return
	}

	// if ctx is done, a new one is created w/ timeout 5s to clean up resources
	if ctx.Err() != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	return e.sandbox.Terminate(ctx)
}

func (e *jobExecutor) initializeJob(s *scribe.Scribe, scope *dig.Scope) error {
	// inject dependencies
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
	e.steps = make(map[string]*records.Step, len(e.jobSpec.Steps))

	// setup expression.Env
	opts := []expression.Option{
		expression.WithVariable("github", &e.github),
		expression.WithVariable("job", e.jobInfo),
		expression.WithVariable("steps", e.steps),
		expression.WithVariable("env", e.env),
		expression.WithLibrary(libraries.StatusLib(e)),
	}
	if exprEnv, err := e.exprEnv.New(opts...); err != nil {
		return err
	} else {
		e.exprEnv = exprEnv
	}

	// Evaluate expressions
	s.Debugf("Evaluating job-level environment variables")
	env := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.jobSpec.Env, &env); err != nil {
		return err
	} else {
		e.SetEnv(env)
	}

	s.Debugf("Evaluating job defaults")
	var defaults workflows.Defaults
	if err := evaluator.Evaluate(e.exprEnv, e.jobSpec.Defaults, &defaults); err != nil {
		return err
	} else if err = xdig.Supply(scope, defaults); err != nil {
		return err
	}

	return nil
}

func (e *jobExecutor) initializeSandbox(ctx context.Context, s *scribe.Scribe, scope *dig.Scope) error {
	var runtime sandboxer.Engine
	if err := xdig.Populate(scope, &runtime); err != nil {
		return err
	}

	req := &sandboxer.LaunchRequest{
		Uid:    JobUid(e),
		Github: &e.github,
	}

	s.Debugf("Evaluating job container")
	var jobContainer *workflows.Container
	if err := evaluator.Evaluate(e.exprEnv, e.jobSpec.Container, &jobContainer); err != nil {
		return err
	} else {
		req.JobContainer = jobContainer
	}

	s.Debugf("Evaluating service containers")
	var serviceContainers = make(map[string]*workflows.Container)
	if err := evaluator.Evaluate(e.exprEnv, e.jobSpec.Services, &serviceContainers); err != nil {
		return err
	} else if len(serviceContainers) > 0 {
		req.ServiceContainers = serviceContainers
	}

	if resp, err := runtime.Launch(ctx, req); err != nil {
		return err
	} else {
		e.sandbox = resp.Sandbox

		// set records values
		s.Debugf("Update context data")
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

	if location, err := e.setupEventFile(ctx); err != nil {
		return err
	} else {
		e.github.EventPath = location
	}

	// register SandboxLib (e.g. hashFiles func) to expression.Env
	var cp xcontext.Provider
	if err := xdig.Populate(scope, &cp); err != nil {
		return err
	}
	opt := expression.WithLibrary(libraries.SandboxLib(cp, e.sandbox))
	if exprEnv, err := e.exprEnv.New(opt); err != nil {
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
	if err := xdig.Supply[support.PathProvider](scope, e); err != nil {
		return err
	}

	return scope.Invoke(e.provideEnv)
}

func (e *jobExecutor) provideEnv(envProv support.EnvProvider) {
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
	envProv.ProvideEnv(support.StaticEnv(runnerEnv))
}

func (e *jobExecutor) initializeSteps(ctx context.Context, scope *dig.Scope) error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.jobSpec.Steps))

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

	for _, step := range e.jobSpec.Steps {
		exec := NewStepExecutor(step)
		e.stepExecutors[step.StepId()] = exec

		s := scope.Scope(fmt.Sprintf("step(%s)", step.StepId()))
		if err := exec.Initialize(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (e *jobExecutor) runStage(ctx context.Context, stage Stage) (ex error) {
	jobId := JobId(e)
	ctx, done := xotel.SetupTelemetry(ctx,
		fmt.Sprintf("JobExecutor.RunStage(%s, %s)", jobId, stage),
		xotel.DrassiStage(stage),
	)
	defer done(&ex)

	// setup listener
	if eh := e.listener.OnRunStage(e, stage); eh != nil {
		err := eh.Begin(ctx)
		defer end(eh, &ex)
		if err != nil {
			return err
		}
	}

	// do stage run
	ids := make([]string, len(e.jobSpec.Steps))
	for i, step := range e.jobSpec.Steps {
		ids[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(ids) // in place reverse
	}
	for _, id := range ids {
		exec := e.stepExecutors[id]
		res := exec.RunStep(ctx, stage)
		if res == nil {
			continue
		}
		// Only set `steps` records in `main` stage & `id` is user specified
		if stage == StageMain && !strings.HasPrefix(id, "__") {
			e.steps[id] = res
		}
		if res.Conclusion == records.ResultFailure {
			e.job.Result = records.ResultFailure
			clog.WarnContextf(ctx, "set job.Result='failure' because of step %s failed", id)
			return fmt.Errorf(`step %q (%s) failed`, id, stage)
		}
	}

	return nil
}

func (e *jobExecutor) State() *records.Job {
	return e.job
}

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

func (e *jobExecutor) SystemPaths() []string {
	return slices.Clone(e.paths)
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

// SetEnv make an environment variable available to any subsequent steps in a workflow job.
//
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *jobExecutor) SetEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

func (e *jobExecutor) setupEventFile(ctx context.Context) (string, error) {
	files := map[string]any{"workflow/event.json": e.github.Event}
	r, err := xtar.JsonObjectReader(files, false)
	if err != nil {
		return "", err
	}

	if err = e.sandbox.CopyIn(ctx, r, e.runner.Temp); err != nil {
		return "", err
	}

	location := path.Join(e.runner.Temp, "workflow", "event.json")
	return location, nil
}
