package executor

import (
	"context"
	"fmt"
	"github.com/dungdm93/drassi/core/pkg/executor/problem"
	"slices"

	"github.com/dungdm93/drassi/core/pkg/container"
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
	JobRun *JobRun

	sandbox       sandboxer.Sandbox
	stepExecutors map[string]*StepExecutor

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

func (e *JobExecutor) Sandbox() sandboxer.Sandbox {
	return e.sandbox
}

func (e *JobExecutor) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	var err error
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
	// TODO
}

func (e *JobExecutor) AddProblemMatcher(owner string, matcher problem.Matcher) {
	// TODO
}

func (e *JobExecutor) RemoveProblemMatcher(owner string) {
	// TODO
}

func (e *JobExecutor) Log(tag, format string, a ...any) {
	// TODO
	message := fmt.Sprintf(format, a...)
	fmt.Println(message)
}

func (e *JobExecutor) toContainerConfig(ctx context.Context, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}
