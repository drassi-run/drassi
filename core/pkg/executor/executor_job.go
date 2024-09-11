package executor

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
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
	RunJob() error
	Finalize() error

	AddPath(paths []string) error
	SetEnv(env map[string]string) error

	// TODO: remove
	ComposePath() string
}

func NewJobExecutor(run *JobRun) JobExecutor {
	return &jobExecutor{
		jobRun: run,
	}
}

type jobExecutor struct {
	jobRun *JobRun

	// records
	gh    dossiers.Github
	job   *dossiers.Job
	steps map[string]*dossiers.Step
	env   map[string]string
	paths []string

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

	return e.initializeSteps(scope)
}

func (e *jobExecutor) RunJob() error {
	if err := e.runStage(StagePre, StepRun.PreTask); err != nil {
		e.job.Status = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(StageMain, StepRun.MainTask); err != nil {
		e.job.Status = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(StagePost, StepRun.PostTask); err != nil {
		e.job.Status = dossiers.ResultFailure
		return err
	}
	e.job.Status = dossiers.ResultSuccess
	return nil
}

func (e *jobExecutor) Finalize() (err error) {
	defer func() {
		output := make(map[string]string)
		if ex := evaluator.Evaluate(e.exprEnv, e.jobRun.Outputs, &output); err == nil && ex != nil {
			err = ex
		}

		e.supervisor.AfterJobRun(e, output)
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
	if err := xdig.Populate(scope, &e.gh); err != nil {
		return err
	} else {
		// sanitize GitHub
		e.gh.Action = ""
		e.gh.ActionPath = ""
		e.gh.ActionRef = ""
		e.gh.ActionRepository = ""
		e.gh.ActionStatus = ""
	}
	e.job = new(dossiers.Job)
	e.steps = make(map[string]*dossiers.Step, len(e.jobRun.Steps))

	e.supervisor.BeforeJobRun(e)

	// setup expression.Env
	opts := []expression.EnvOption{
		expression.WithVariable("github", &e.gh),
		expression.WithVariable("job", e.job),
		expression.WithVariable("steps", e.steps),
		expression.WithVariable("env", e.env),
		expression.WithLibrary(libraries.JobLib(e.job)),
	}
	if exprEnv, err := expression.NewEnv(e.exprEnv, opts...); err != nil {
		return err
	} else {
		e.exprEnv = exprEnv
	}

	// Provide scope values
	if err := xdig.Supply(scope, e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.gh); err != nil {
		return err
	}
	if err := xdig.Supply(scope, e.env); err != nil {
		return err
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
		e.job.Container = resp.Container
		e.job.Services = resp.Services
	}

	if err := xdig.Supply(scope, e.sandbox, dig.Export(true)); err != nil {
		return err
	}

	return nil
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

func (e *jobExecutor) ComposePath() string {
	return strings.Join(e.paths, ":")
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

// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
//func (e *jobExecutor) ciEnv() map[string]string {
//	gh := e.dossier.Github
//	r := e.dossier.Runner
//
//	m := map[string]string{
//		"CI":             "true",
//		"GITHUB_ACTIONS": "true",
//
//		"GITHUB_ACTOR":               gh.Actor,
//		"GITHUB_ACTOR_ID":            gh.ActorId,
//		"GITHUB_API_URL":             gh.ApiUrl,
//		"GITHUB_BASE_REF":            gh.BaseRef,
//		"GITHUB_EVENT_NAME":          gh.EventName,
//		"GITHUB_GRAPHQL_URL":         gh.GraphqlUrl,
//		"GITHUB_HEAD_REF":            gh.HeadRef,
//		"GITHUB_JOB":                 gh.Job,
//		"GITHUB_REF":                 gh.Ref,
//		"GITHUB_REF_NAME":            gh.RefName,
//		"GITHUB_REF_PROTECTED":       strconv.FormatBool(gh.RefProtected),
//		"GITHUB_REF_TYPE":            string(gh.RefType),
//		"GITHUB_REPOSITORY":          gh.Repository,
//		"GITHUB_REPOSITORY_ID":       gh.RepositoryId,
//		"GITHUB_REPOSITORY_OWNER":    gh.RepositoryOwner,
//		"GITHUB_REPOSITORY_OWNER_ID": gh.RepositoryOwnerId,
//		"GITHUB_RETENTION_DAYS":      gh.RetentionDays,
//		"GITHUB_RUN_ATTEMPT":         gh.RunAttempt,
//		"GITHUB_RUN_ID":              gh.RunId,
//		"GITHUB_RUN_NUMBER":          gh.RunNumber,
//		"GITHUB_SERVER_URL":          gh.ServerUrl,
//		"GITHUB_SHA":                 gh.Sha,
//		"GITHUB_TRIGGERING_ACTOR":    gh.TriggeringActor,
//		"GITHUB_WORKFLOW":            gh.Workflow,
//		"GITHUB_WORKFLOW_REF":        gh.WorkflowRef,
//		"GITHUB_WORKFLOW_SHA":        gh.WorkflowSha,
//
//		"RUNNER_NAME":        r.Name,
//		"RUNNER_ARCH":        string(r.Arch),
//		"RUNNER_OS":          string(r.Os),
//		"RUNNER_ENVIRONMENT": r.Environment,
//	}
//	if r.Debug == "1" {
//		m["RUNNER_DEBUG"] = r.Debug
//	}
//
//	return m
//}

//func (e *jobExecutor) processSandboxEnv(env map[string]string) {
//	gh := e.dossier.Github
//	gh.Workspace = env["GITHUB_WORKSPACE"]
//	gh.EventPath = env["GITHUB_EVENT_PATH"]
//
//	r := e.dossier.Runner
//	r.Temp = env["RUNNER_TEMP"]
//	r.ToolCache = env["RUNNER_TOOL_CACHE"]
//	r.Workspace = env["RUNNER_WORKSPACE"]
//	// env.RUNNER_USER
//	// env.RUNNER_PERFLOG
//
//	e.paths = strings.Split(env["PATH"], ":")
//}

//// File commands env
// "GITHUB_PATH": gh.Path
// "GITHUB_ENV": gh.Env
// "GITHUB_OUTPUT": gh.Output
// "GITHUB_STATE": gh.State
// "GITHUB_STEP_SUMMARY": gh.StepSummary
