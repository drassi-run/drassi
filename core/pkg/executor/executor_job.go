package executor

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/container"
	"github.com/dungdm93/drassi/core/pkg/executor/command"
	"github.com/dungdm93/drassi/core/pkg/executor/problem"
	"github.com/dungdm93/drassi/core/pkg/executor/reporter"
	"github.com/dungdm93/drassi/core/pkg/executor/secret"
	"github.com/dungdm93/drassi/core/pkg/model/contexts"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/core/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	setEnvBlockList = sets.New("NODE_OPTIONS")
)

type JobExecutor struct {
	JobRun   *JobRun
	Reporter reporter.Reporter

	sandbox       sandboxer.Sandbox
	stepExecutors map[string]*StepExecutor

	secretMasker    secret.Masker
	problemMatchers map[string]problem.Matcher

	outWriter         io.Writer
	errWriter         io.Writer
	streams           *sandboxer.Streams
	consoleCmdMgr     *command.ConsoleCommandManager
	consoleCmdHandler *consoleCommandHandlers

	defaults workflows.Defaults
	env      map[string]string
	paths    []string
}

func (e *JobExecutor) NewStepExecutor(step StepRun) *StepExecutor {
	exec := &StepExecutor{
		job:         e,
		parent:      nil,
		children:    make(map[string]*StepExecutor),
		stepRun:     step,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &contexts.Step{},
	}
	e.stepExecutors[step.StepId()] = exec
	return exec
}

func (e *JobExecutor) StepExecutor(id string) *StepExecutor {
	return e.stepExecutors[id]
}

func (e *JobExecutor) ContextData(name string) context.Context {
	return context.Background()
}

func (e *JobExecutor) Functions(name string) []string {
	return nil
}

func (e *JobExecutor) DefaultValue(name string) any {
	return nil
}

func (e *JobExecutor) Streams() *sandboxer.Streams {
	return e.streams
}

func (e *JobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *JobExecutor) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	var err error
	if err = e.initializeJob(ctx); err != nil {
		return err
	}

	e.env, err = e.JobRun.Env.Evaluate("job.env", e)
	if err != nil {
		return err
	}

	e.defaults, err = e.JobRun.Defaults.Evaluate("job.defaults", e)
	if err != nil {
		return err
	}

	if err = e.initializeSandbox(ctx, runtime); err != nil {
		return err
	}

	return e.initializeSteps(ctx)
}

func (e *JobExecutor) RunJob(ctx context.Context) error {
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

func (e *JobExecutor) Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	req := sandboxer.TerminateSandboxRequest{
		Sandbox: e.sandbox,
	}
	_, err := runtime.TerminateSandbox(ctx, req)
	return err
}

func (e *JobExecutor) initializeJob(ctx context.Context) error {
	e.outWriter = secret.NewWriter(e.Reporter.Stdout(), &e.secretMasker)
	e.errWriter = secret.NewWriter(e.Reporter.Stderr(), &e.secretMasker)

	lineOutWriter := reporter.NewLineWriter(e.lineHandler(e.outWriter))
	lineErrWriter := reporter.NewLineWriter(e.lineHandler(e.errWriter))

	e.streams = &sandboxer.Streams{
		In:  e.Reporter.Stdin(),
		Out: lineOutWriter,
		Err: lineErrWriter,
	}

	e.consoleCmdMgr = command.NewConsoleCommandManager(e.outWriter)
	e.consoleCmdHandler = &consoleCommandHandlers{e.consoleCmdMgr}
	e.consoleCmdHandler.RegisterForJobExecutor(ctx, e)
	return nil
}

func (e *JobExecutor) initializeSandbox(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	var jobContainer *container.ContainerConfig
	if con, err := e.JobRun.Container.Evaluate("job.container", e); err != nil {
		return err
	} else {
		jobContainer, err = e.toContainerConfig(ctx, con)
		if err != nil {
			return err
		}
	}

	var serviceContainers = make(map[string]*container.ContainerConfig)
	if sers, err := e.JobRun.Services.Evaluate("job.services", e); err != nil {
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
		JobId:             e.JobRun.ID,
		JobEnv:            e.env,
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if res, err := runtime.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		e.sandbox = res.Sandbox
	}
	return nil
}

func (e *JobExecutor) initializeSteps(ctx context.Context) error {
	e.stepExecutors = make(map[string]*StepExecutor, len(e.JobRun.Steps))
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
