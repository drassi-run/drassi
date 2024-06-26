package executor

import (
	"context"
	"fmt"
	"io"
	"maps"
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

var (
	setEnvBlockList = sets.New("NODE_OPTIONS")
)

type JobExecutor struct {
	JobRun   *JobRun
	Reporter reporter.Reporter
	Dossier  *dossiers.Dossier

	sandbox       sandboxer.Sandbox
	stepExecutors map[string]StepExecutor

	secretMasker    secret.Masker
	problemMatchers map[string]problem.Matcher

	outWriter         io.Writer
	errWriter         io.Writer
	streams           *sandboxer.Streams
	consoleCmdMgr     *command.ConsoleCommandManager
	consoleCmdHandler *consoleCommandHandlers

	defaults workflows.Defaults
	env      map[string]string // ref Dossier.Env
	paths    []string
	result   dossiers.Result
}

func (e *JobExecutor) JobId() string {
	return e.JobRun.Id
}

func (e *JobExecutor) NewStepExecutor(step StepRun) StepExecutor {
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

func (e *JobExecutor) StepExecutor(id string) StepExecutor {
	return e.stepExecutors[id]
}

func (e *JobExecutor) NewSubDossier() *dossiers.Dossier {
	// Github context is cloned because `github.action_*` can be set by step
	gh := *e.Dossier.Github // shallow clone GitHub

	// env context is cloned because of step level env
	env := maps.Clone(e.Dossier.Env)

	return &dossiers.Dossier{
		Github:    &gh,
		Env:       env,
		Variables: e.Dossier.Variables,
		Job:       e.Dossier.Job,
		Jobs:      e.Dossier.Jobs,
		Steps:     e.Dossier.Steps,
		Runner:    e.Dossier.Runner,
		Secrets:   e.Dossier.Secrets, // TODO: secrets context is not available for composite actions
		Strategy:  e.Dossier.Strategy,
		Matrix:    e.Dossier.Matrix,
		Needs:     e.Dossier.Needs,
		Inputs:    e.Dossier.Inputs,
	}
}

func (e *JobExecutor) Streams() *sandboxer.Streams {
	return e.streams
}

func (e *JobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *JobExecutor) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	e.Reporter.StartJob()

	if err := e.initializeJob(ctx); err != nil {
		return err
	}

	if err := e.initializeSandbox(ctx, runtime); err != nil {
		return err
	}

	return e.initializeSteps(ctx)
}

func (e *JobExecutor) RunJob(ctx context.Context) error {
	if err := e.consoleCmdHandler.StartJob(ctx, e); err != nil {
		return err
	}
	defer e.consoleCmdHandler.EndJob()

	if err := e.runStage(ctx, StagePre, StepRun.PreTask); err != nil {
		return err
	}
	if err := e.runStage(ctx, StageMain, StepRun.MainTask); err != nil {
		return err
	}
	if err := e.runStage(ctx, StagePost, StepRun.PostTask); err != nil {
		return err
	}
	return nil
}

func (e *JobExecutor) Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) (err error) {
	evalSupplier := &evaluationSupplier{dossier: e.Dossier}
	defer func() {
		output, ex := e.JobRun.Outputs.Evaluate("job.output", evalSupplier)
		if err != nil && ex != nil {
			err = ex
		}

		e.Reporter.EndJob(e.result, output)
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

func (e *JobExecutor) initializeJob(ctx context.Context) error {
	e.outWriter = secret.NewWriter(e.Reporter.Stdout(), e.secretMasker)
	e.errWriter = secret.NewWriter(e.Reporter.Stderr(), e.secretMasker)

	lineOutWriter := reporter.NewLineWriter(e.lineHandler(e.outWriter))
	lineErrWriter := reporter.NewLineWriter(e.lineHandler(e.errWriter))

	e.streams = &sandboxer.Streams{
		In:  e.Reporter.Stdin(),
		Out: lineOutWriter,
		Err: lineErrWriter,
	}

	e.consoleCmdMgr = command.NewConsoleCommandManager(e.outWriter)
	e.consoleCmdHandler = &consoleCommandHandlers{cmdMgr: e.consoleCmdMgr}
	evalSupplier := &evaluationSupplier{dossier: e.Dossier}

	e.sanitizeDossier()
	e.env = e.Dossier.Env
	if env, err := e.JobRun.Env.Evaluate("job.env", evalSupplier); err != nil {
		return err
	} else {
		if err = e.SetEnv(env); err != nil {
			return err
		}
	}

	if defaults, err := e.JobRun.Defaults.Evaluate("job.defaults", evalSupplier); err != nil {
		return err
	} else {
		e.defaults = defaults
	}
	return nil
}

func (e *JobExecutor) initializeSandbox(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	evalSupplier := &evaluationSupplier{dossier: e.Dossier}
	var jobContainer *container.ContainerConfig
	if con, err := e.JobRun.Container.Evaluate("job.container", evalSupplier); err != nil {
		return err
	} else {
		jobContainer, err = e.toContainerConfig(ctx, con)
		if err != nil {
			return err
		}
	}

	var serviceContainers = make(map[string]*container.ContainerConfig)
	if sers, err := e.JobRun.Services.Evaluate("job.services", evalSupplier); err != nil {
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
		JobId:             e.JobRun.Id,
		JobEnv:            e.ciEnv(),
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if resp, err := runtime.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		e.sandbox = resp.Sandbox
		e.Dossier.Job.Container = resp.Container
		e.Dossier.Job.Services = resp.Services

		e.processSandboxEnv(resp.Env)
	}
	return nil
}

func (e *JobExecutor) initializeSteps(ctx context.Context) error {
	e.stepExecutors = make(map[string]StepExecutor, len(e.JobRun.Steps))
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range e.JobRun.Steps {
		exec := e.NewStepExecutor(step)
		g.Go(func() error {
			return exec.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (e *JobExecutor) runStage(ctx context.Context, stage Stage, fn func(StepRun) *Task) error {
	ids := make([]string, len(e.JobRun.Steps))
	for i, step := range e.JobRun.Steps {
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

// Add paths to the context and remove duplicates
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path
func (e *JobExecutor) AddPath(paths []string) error {
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

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (e *JobExecutor) SetEnv(env map[string]string) error {
	for k, v := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			continue
		}
		e.env[k] = v
	}
	return nil
}

func (e *JobExecutor) AddSecretMask(secret secret.Secret) {
	e.secretMasker.AddSecret(secret)
}

func (e *JobExecutor) AddProblemMatcher(owner string, matcher problem.Matcher) {
	if e.problemMatchers == nil {
		e.problemMatchers = make(map[string]problem.Matcher)
	}
	e.problemMatchers[owner] = matcher
}

func (e *JobExecutor) RemoveProblemMatcher(owner string) {
	delete(e.problemMatchers, owner)
}

// https://en.wikipedia.org/wiki/ANSI_escape_code
var colorCodeRegex = regexp.MustCompile(`\033\[[\d;]*m`)

func (e *JobExecutor) scanProblem(line string) error {
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
		return e.Reporter.AddIssue(issue)
	}
}

func (e *JobExecutor) lineHandler(w io.Writer) reporter.LineHandler {
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

func (e *JobExecutor) Log(tag, format string, a ...any) {
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

func (e *JobExecutor) sanitizeDossier() {
	d := e.Dossier
	gh := d.Github

	gh.Action = ""
	gh.ActionPath = ""
	gh.ActionRef = ""
	gh.ActionRepository = ""
	gh.ActionStatus = ""

	d.Job = new(dossiers.Job)

	if d.Env == nil {
		e.Dossier.Env = make(map[string]string)
	}
	if d.Steps == nil {
		d.Steps = make(map[string]*dossiers.Step, len(e.JobRun.Steps))
	}
}

// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
func (e *JobExecutor) ciEnv() map[string]string {
	gh := e.Dossier.Github
	r := e.Dossier.Runner

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

func (e *JobExecutor) processSandboxEnv(env map[string]string) {
	gh := e.Dossier.Github
	gh.Workspace = env["GITHUB_WORKSPACE"]
	gh.EventPath = env["GITHUB_EVENT_PATH"]

	r := e.Dossier.Runner
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
