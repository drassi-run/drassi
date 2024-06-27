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
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

type JobExecutor interface {
	JobId() string
	Streams() *sandboxer.Streams
	Sandbox() sandboxer.Sandbox
	Reporter() reporter.Reporter
	NewSubDossier() *dossiers.Dossier

	Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error
	RunJob(ctx context.Context) error
	Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) error
	Defaults() *workflows.Defaults
	ComposePath() string

	StartStep(ctx context.Context, stepExec StepExecutor) error
	EndStep()

	Log(tag, format string, a ...any)
	AddPath(paths []string) error
	SetEnv(env map[string]string) error
	AddSecretMask(secret secret.Secret)
	AddProblemMatcher(owner string, matcher problem.Matcher)
	RemoveProblemMatcher(owner string)
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

	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor

	secretMasker    secret.Masker
	problemMatchers map[string]problem.Matcher

	outWriter     io.Writer
	errWriter     io.Writer
	streams       *sandboxer.Streams
	consoleCmdMgr command.ConsoleCommandManager
	fileCmdMgr    command.FileCommandManager
	cmdHandlers   *commandHandlers

	defaults *workflows.Defaults
	paths    []string
	result   dossiers.Result
}

func (e *jobExecutor) JobId() string {
	return e.jobRun.Id
}

func (e *jobExecutor) NewStepExecutor(step StepRun) StepExecutor {
	exec := &stepExecutor{
		job:      e,
		parent:   nil,
		children: make(map[string]StepExecutor),
		stepRun:  step,
		state:    make(map[string]string),
		result: &dossiers.Step{
			Outputs: make(map[string]string),
		},
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

func (e *jobExecutor) Reporter() reporter.Reporter {
	return e.reporter
}

func (e *jobExecutor) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	e.reporter.StartJob()

	if err := e.initializeJob(ctx); err != nil {
		return err
	}

	if err := e.initializeSandbox(ctx, runtime); err != nil {
		return err
	}

	return e.initializeSteps(ctx)
}

func (e *jobExecutor) RunJob(ctx context.Context) error {
	if err := e.cmdHandlers.StartJob(ctx, e); err != nil {
		return err
	}
	defer e.cmdHandlers.EndJob()

	if err := e.runStage(ctx, StagePre, StepRun.PreTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(ctx, StageMain, StepRun.MainTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	if err := e.runStage(ctx, StagePost, StepRun.PostTask); err != nil {
		e.result = dossiers.ResultFailure
		return err
	}
	e.result = dossiers.ResultSuccess
	return nil
}

func (e *jobExecutor) Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) (err error) {
	evalSupplier := &evaluationSupplier{dossier: e.dossier}
	defer func() {
		output, ex := e.jobRun.Outputs.Evaluate("job.output", evalSupplier)
		if err != nil && ex != nil {
			err = ex
		}

		e.reporter.EndJob(e.result, output)
	}()

	if e.sandbox == nil {
		return
	}

	// if ctx is done, a new one is created w/ timeout 5s to clean-up resources
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

func (e *jobExecutor) initializeJob(ctx context.Context) error {
	e.outWriter = secret.NewWriter(e.reporter.Stdout(), e.secretMasker)
	e.errWriter = secret.NewWriter(e.reporter.Stderr(), e.secretMasker)
	e.streams = &sandboxer.Streams{
		In:  e.reporter.Stdin(),
		Out: reporter.NewLineWriter(e.lineHandler(e.outWriter)),
		Err: reporter.NewLineWriter(e.lineHandler(e.errWriter)),
	}

	e.consoleCmdMgr = command.NewConsoleCommandManager(e.outWriter)
	e.fileCmdMgr = command.NewFileCommandManager(e.jobRun.Uid)
	e.cmdHandlers = &commandHandlers{
		consoleMgr: e.consoleCmdMgr,
		fileMgr:    e.fileCmdMgr,
	}

	// Evaluate expressions
	evalSupplier := &evaluationSupplier{dossier: e.dossier}

	if env, err := e.jobRun.Env.Evaluate("job.env", evalSupplier); err != nil {
		return err
	} else {
		if err = e.SetEnv(env); err != nil {
			return err
		}
	}

	if defaults, err := e.jobRun.Defaults.Evaluate("job.defaults", evalSupplier); err != nil {
		return err
	} else {
		e.defaults = &defaults
	}
	return nil
}

func (e *jobExecutor) initializeSandbox(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	evalSupplier := &evaluationSupplier{dossier: e.dossier}
	var jobContainer *container.ContainerConfig
	if con, err := e.jobRun.Container.Evaluate("job.container", evalSupplier); err != nil {
		return err
	} else {
		jobContainer, err = e.toContainerConfig(ctx, con)
		if err != nil {
			return err
		}
	}

	var serviceContainers = make(map[string]*container.ContainerConfig)
	if sers, err := e.jobRun.Services.Evaluate("job.services", evalSupplier); err != nil {
		return err
	} else {
		for name, ser := range sers {
			con, err := e.toContainerConfig(ctx, ser)
			if err != nil {
				return err
			}
			serviceContainers[name] = con
		}
	}

	req := sandboxer.LaunchSandboxRequest{
		JobId:             e.jobRun.Id,
		JobEnv:            e.ciEnv(),
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if resp, err := runtime.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		e.sandbox = resp.Sandbox
		e.dossier.Job.Container = resp.Container
		e.dossier.Job.Services = resp.Services

		e.processSandboxEnv(resp.Env)
	}
	return nil
}

func (e *jobExecutor) initializeSteps(ctx context.Context) error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.jobRun.Steps))
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range e.jobRun.Steps {
		exec := e.NewStepExecutor(step)
		g.Go(func() error {
			return exec.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (e *jobExecutor) runStage(ctx context.Context, stage Stage, fn func(StepRun) *Task) error {
	ids := make([]string, len(e.jobRun.Steps))
	for i, step := range e.jobRun.Steps {
		ids[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(ids) // in place reverse
	}
	for _, id := range ids {
		exec := e.StepExecutor(id)
		if err := exec.RunStep(ctx, fn); err != nil {
			return err
		}
	}

	return nil
}

func (e *jobExecutor) StartStep(ctx context.Context, stepExec StepExecutor) error {
	return e.cmdHandlers.StartStep(ctx, stepExec)
}

func (e *jobExecutor) EndStep() {
	e.cmdHandlers.EndStep()
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

func (e *jobExecutor) lineHandler(w io.Writer) reporter.LineHandler {
	return func(line string) error {
		if cmd := e.consoleCmdMgr.ParseCommand(line); cmd != nil {
			if err := e.consoleCmdMgr.Process(line, cmd); err != nil {
				e.Log(TagError, err.Error())
			}
			return nil
		}
		if err := e.scanProblem(line); err != nil {
			e.Log(TagError, err.Error())
		}
		_, err := io.WriteString(w, line)
		return err
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
