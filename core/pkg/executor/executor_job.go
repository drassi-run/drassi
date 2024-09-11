package executor

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"drassi.run/core/pkg/container"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobCommandHandler interface {
	JobRun() *JobRun
	Sandbox() sandboxer.Sandbox
	Context() context.Context

	Log(tag, format string, a ...any)
	AddPath(paths []string) error
	SetEnv(env map[string]string) error
	AddSecretMask(secret secret.Secret)
	AddProblemMatcher(owner string, matcher problem.Matcher)
	RemoveProblemMatcher(owner string)
}

type JobExecutor interface {
	JobCommandHandler

	JobId() string
	Streams() *sandboxer.Streams
	Sandbox() sandboxer.Sandbox
	NewSubDossier() *dossiers.Dossier

	SetContext(ctx context.Context)

	Initialize(runtime sandboxer.SandboxRuntime) error
	RunJob() error
	Finalize(runtime sandboxer.SandboxRuntime) error

	Defaults() *workflows.Defaults
	ComposePath() string
}

func NewJobExecutor(run *JobRun, d *dossiers.Dossier, rep reporter.Reporter) JobExecutor {
	// sanitize dossier
	gh := d.Github
	gh.Action = ""
	gh.ActionPath = ""
	gh.ActionRef = ""
	gh.ActionRepository = ""
	gh.ActionStatus = ""

	d.Job = new(dossiers.Job)
	if d.Env == nil {
		d.Env = make(map[string]string)
	}
	if d.Steps == nil {
		d.Steps = make(map[string]*dossiers.Step, len(run.Steps))
	}

	je := &jobExecutor{
		jobRun:   run,
		dossier:  d,
		reporter: rep,

		secretMasker:    secret.NewMasker(),
		problemMatchers: make(map[string]problem.Matcher),

		// will be overwritten in initializeJob
		outWriter: os.Stdout,
		errWriter: os.Stderr,
	}

	return je
}

type jobExecutor struct {
	jobRun   *JobRun
	dossier  *dossiers.Dossier
	reporter reporter.Reporter

	ctx context.Context

	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor

	secretMasker    secret.Masker
	problemMatchers map[string]problem.Matcher

	outWriter io.Writer
	errWriter io.Writer
	streams   *sandboxer.Streams
	cmdCtrl   CommandController
	exprEnv   *expression.Env

	defaults *workflows.Defaults
	paths    []string
	result   dossiers.Result
}

func (e *jobExecutor) JobId() string {
	return e.jobRun.Id
}

func (e *jobExecutor) JobRun() *JobRun {
	return e.jobRun
}

func (e *jobExecutor) NewStepExecutor(step StepRun) StepExecutor {
	exec := &stepExecutor{
		job:      e,
		parent:   nil,
		children: make(map[string]StepExecutor),
		stepRun:  step,
		reporter: e.reporter,
		cmdCtrl:  e.cmdCtrl,
		state:    make(map[string]string),
	}
	e.stepExecutors[step.StepId()] = exec
	return exec
}

func (e *jobExecutor) StepExecutor(id string) StepExecutor {
	return e.stepExecutors[id]
}

func (e *jobExecutor) NewSubDossier() *dossiers.Dossier {
	// GitHub context is cloned because `github.action_*` can be set by step
	gh := *e.dossier.Github // shallow clone GitHub

	// env context is cloned because of step level env
	env := maps.Clone(e.dossier.Env)

	return &dossiers.Dossier{
		Github:    &gh,
		Env:       env,
		Variables: e.dossier.Variables,
		Job:       e.dossier.Job,
		Jobs:      e.dossier.Jobs,
		Steps:     e.dossier.Steps,
		Runner:    e.dossier.Runner,
		Secrets:   e.dossier.Secrets, // TODO: secrets context is not available for composite actions
		Strategy:  e.dossier.Strategy,
		Matrix:    e.dossier.Matrix,
		Needs:     e.dossier.Needs,
		Inputs:    e.dossier.Inputs,
	}
}

func (e *jobExecutor) Streams() *sandboxer.Streams {
	return e.streams
}

func (e *jobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *jobExecutor) Context() context.Context {
	return e.ctx
}

func (e *jobExecutor) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *jobExecutor) Initialize(runtime sandboxer.SandboxRuntime) error {
	e.reporter.StartJob()

	if err := e.initializeJob(); err != nil {
		return err
	}

	if err := e.initializeSandbox(runtime); err != nil {
		return err
	}

	return e.initializeSteps()
}

func (e *jobExecutor) RunJob() error {
	e.cmdCtrl.Register()

	if err := e.runStage(StagePre, StepRun.PreTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(StageMain, StepRun.MainTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(StagePost, StepRun.PostTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	e.result = dossiers.ResultSuccess
	return nil
}

func (e *jobExecutor) Finalize(runtime sandboxer.SandboxRuntime) (err error) {
	defer func() {
		output := make(map[string]string)
		if ex := evaluator.Evaluate(e.exprEnv, e.jobRun.Outputs, &output); err == nil && ex != nil {
			err = ex
		}

		e.reporter.EndJob(e.result, output)
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
	_, err = runtime.TerminateSandbox(ctx, req)
	return
}

func (e *jobExecutor) initializeJob() error {
	e.outWriter = secret.NewWriter(e.reporter.Stdout(), e.secretMasker)
	e.errWriter = secret.NewWriter(e.reporter.Stderr(), e.secretMasker)

	e.cmdCtrl = &commandController{
		consoleMgr: command.NewConsoleManager(e.outWriter),
		fileMgr:    command.NewFileManager(e.sandbox),
		job:        e,
	}

	e.streams = &sandboxer.Streams{
		In: e.reporter.Stdin(),
		Out: logging.NewLineWriter(
			e.cmdCtrl.LineHandler(e.outWriter, e.scanProblem),
		),
		Err: logging.NewLineWriter(
			e.cmdCtrl.LineHandler(e.errWriter, e.scanProblem),
		),
	}

	// Evaluate expressions
	env := make(map[string]string)
	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Env, &env); err != nil {
		return err
	} else if err = e.SetEnv(env); err != nil {
		return err
	}

	if err := evaluator.Evaluate(e.exprEnv, e.jobRun.Defaults, e.defaults); err != nil {
		return err
	}

	return nil
}

func (e *jobExecutor) initializeSandbox(runtime sandboxer.SandboxRuntime) error {
	req := sandboxer.LaunchSandboxRequest{
		JobId:  e.jobRun.Id,
		JobEnv: e.ciEnv(),
	}

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

	resp, err := runtime.LaunchSandbox(e.ctx, req)
	if err != nil {
		return err
	}

	e.sandbox = resp.Sandbox
	e.dossier.Job.Container = resp.Container
	e.dossier.Job.Services = resp.Services
	e.processSandboxEnv(resp.Env)
	return nil
}

func (e *jobExecutor) initializeSteps() error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.jobRun.Steps))
	g, ctx := errgroup.WithContext(e.ctx)
	for _, step := range e.jobRun.Steps {
		exec := e.NewStepExecutor(step)
		g.Go(func() error {
			exec.SetContext(ctx)
			return exec.Initialize()
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
		e.dossier.Steps[id] = res
		if res.Conclusion == dossiers.ResultFailure {
			return fmt.Errorf(`step %q (%s) failed`, id, stage)
		}
	}

	return nil
}

func (e *jobExecutor) Defaults() *workflows.Defaults {
	return e.defaults
}

func (e *jobExecutor) ComposePath() string {
	return strings.Join(e.paths, ":")
}

// Add paths to the context and remove duplicates
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

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *jobExecutor) SetEnv(env map[string]string) error {
	for k, v := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			continue
		}
		e.dossier.Env[k] = v
	}
	return nil
}

func (e *jobExecutor) AddSecretMask(secret secret.Secret) {
	e.secretMasker.AddSecret(secret)
}

func (e *jobExecutor) AddProblemMatcher(owner string, matcher problem.Matcher) {
	e.problemMatchers[owner] = matcher
}

func (e *jobExecutor) RemoveProblemMatcher(owner string) {
	delete(e.problemMatchers, owner)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (e *jobExecutor) scanProblem(line string) error {
	line = colorCodeRegex.ReplaceAllLiteralString(line, "")
	var owner string
	var pbl *problem.Problem

	for o, m := range e.problemMatchers {
		if p := m.Match(line); p != nil {
			owner = o
			pbl = p
			break
		}
	}

	// Not matched
	if pbl == nil {
		return nil
	}

	// Matched
	// 1. Reset other matchers
	for o, m := range e.problemMatchers {
		if o != owner {
			m.Reset()
		}
	}

	// 2. convert Problem to Issue
	if issue, err := e.toIssuer(pbl); err != nil {
		return err
	} else {
		// 3. Report the issue
		return e.reporter.AddIssue(issue)
	}
}

const (
	TagGroup    = "##[group]"
	TagEndGroup = "##[endgroup]"
	TagSection  = "##[section]"
	TagCommand  = "##[command]"
	TagError    = "##[error]"
	TagWarning  = "##[warning]"
	TagNotice   = "##[notice]"
	TagDebug    = "##[debug]"
)

func (e *jobExecutor) Log(tag, format string, a ...any) {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	if tag != "" {
		message = tag + message
	}
	if !strings.HasSuffix(message, "\n") {
		message = message + "\n"
	}
	_, _ = io.WriteString(e.outWriter, message)
}

// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
func (e *jobExecutor) ciEnv() map[string]string {
	gh := e.dossier.Github
	r := e.dossier.Runner

	m := map[string]string{
		"CI":             "true",
		"GITHUB_ACTIONS": "true",

		"GITHUB_ACTOR":               gh.Actor,
		"GITHUB_ACTOR_ID":            gh.ActorId,
		"GITHUB_API_URL":             gh.ApiUrl,
		"GITHUB_BASE_REF":            gh.BaseRef,
		"GITHUB_EVENT_NAME":          gh.EventName,
		"GITHUB_GRAPHQL_URL":         gh.GraphqlUrl,
		"GITHUB_HEAD_REF":            gh.HeadRef,
		"GITHUB_JOB":                 gh.Job,
		"GITHUB_REF":                 gh.Ref,
		"GITHUB_REF_NAME":            gh.RefName,
		"GITHUB_REF_PROTECTED":       strconv.FormatBool(gh.RefProtected),
		"GITHUB_REF_TYPE":            string(gh.RefType),
		"GITHUB_REPOSITORY":          gh.Repository,
		"GITHUB_REPOSITORY_ID":       gh.RepositoryId,
		"GITHUB_REPOSITORY_OWNER":    gh.RepositoryOwner,
		"GITHUB_REPOSITORY_OWNER_ID": gh.RepositoryOwnerId,
		"GITHUB_RETENTION_DAYS":      gh.RetentionDays,
		"GITHUB_RUN_ATTEMPT":         gh.RunAttempt,
		"GITHUB_RUN_ID":              gh.RunId,
		"GITHUB_RUN_NUMBER":          gh.RunNumber,
		"GITHUB_SERVER_URL":          gh.ServerUrl,
		"GITHUB_SHA":                 gh.Sha,
		"GITHUB_TRIGGERING_ACTOR":    gh.TriggeringActor,
		"GITHUB_WORKFLOW":            gh.Workflow,
		"GITHUB_WORKFLOW_REF":        gh.WorkflowRef,
		"GITHUB_WORKFLOW_SHA":        gh.WorkflowSha,

		"RUNNER_NAME":        r.Name,
		"RUNNER_ARCH":        string(r.Arch),
		"RUNNER_OS":          string(r.Os),
		"RUNNER_ENVIRONMENT": r.Environment,
	}
	if r.Debug == "1" {
		m["RUNNER_DEBUG"] = r.Debug
	}

	return m
}

func (e *jobExecutor) processSandboxEnv(env map[string]string) {
	gh := e.dossier.Github
	gh.Workspace = env["GITHUB_WORKSPACE"]
	gh.EventPath = env["GITHUB_EVENT_PATH"]

	r := e.dossier.Runner
	r.Temp = env["RUNNER_TEMP"]
	r.ToolCache = env["RUNNER_TOOL_CACHE"]
	r.Workspace = env["RUNNER_WORKSPACE"]
	// env.RUNNER_USER
	// env.RUNNER_PERFLOG

	e.paths = strings.Split(env["PATH"], ":")
}

//// File commands env
// "GITHUB_PATH": gh.Path
// "GITHUB_ENV": gh.Env
// "GITHUB_OUTPUT": gh.Output
// "GITHUB_STATE": gh.State
// "GITHUB_STEP_SUMMARY": gh.StepSummary
