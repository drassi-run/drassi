package executor

import (
	"context"
	"slices"

	"github.com/dungdm93/drassi/core/pkg/container"
	"github.com/dungdm93/drassi/core/pkg/model/workflows"
	"github.com/dungdm93/drassi/core/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	setEnvBlockList = sets.New("NODE_OPTIONS")
)

type JobRunContext struct {
	JobRun *JobRun

	sandbox      sandboxer.Sandbox
	stepContexts map[string]*StepRunContext

	defaults workflows.Defaults
	env      map[string]string
	paths    []string
}

func (c *JobRunContext) ContextData(name string) context.Context {
	panic("implement me")
}

func (c *JobRunContext) Functions(name string) []string {
	panic("implement me")
}

func (c *JobRunContext) DefaultValue(name string) any {
	panic("implement me")
}

func (c *JobRunContext) Sandbox() sandboxer.Sandbox {
	return c.sandbox
}

func (c *JobRunContext) Initialize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	var err error
	c.env, err = c.JobRun.Env.Evaluate("job.env", c)
	if err != nil {
		return err
	}

	c.defaults, err = c.JobRun.Defaults.Evaluate("job.defaults", c)
	if err != nil {
		return err
	}

	var jobContainer *container.ContainerConfig
	if con, err := c.JobRun.Container.Evaluate("job.container", c); err != nil {
		return err
	} else {
		jobContainer, err = c.toContainerConfig(ctx, con)
		if err != nil {
			return err
		}
	}
	var serviceContainers = make(map[string]*container.ContainerConfig)
	if sers, err := c.JobRun.Services.Evaluate("job.services", c); err != nil {
		return err
	} else {
		for name, ser := range sers {
			con, err := c.toContainerConfig(ctx, ser)
			if err != nil {
				return err
			}
			serviceContainers[name] = con
		}
	}

	req := sandboxer.LaunchSandboxRequest{
		JobId:             c.JobRun.ID,
		JobEnv:            c.env,
		JobContainer:      jobContainer,
		ServiceContainers: serviceContainers,
	}
	if res, err := runtime.LaunchSandbox(ctx, req); err != nil {
		return err
	} else {
		c.sandbox = res.Sandbox
	}
	return nil
}

func (c *JobRunContext) RunJob(ctx context.Context) error {
	c.makeStepRunContexts()

	if err := c.initializeSteps(ctx); err != nil {
		return err
	}

	if err := c.runStage(ctx, StagePre, StepRun.PreTask); err != nil {
		return err
	}
	if err := c.runStage(ctx, StageMain, StepRun.MainTask); err != nil {
		return err
	}
	if err := c.runStage(ctx, StagePost, StepRun.PostTask); err != nil {
		return err
	}
	return nil
}

func (c *JobRunContext) Finalize(ctx context.Context, runtime sandboxer.SandboxRuntime) error {
	req := sandboxer.TerminateSandboxRequest{
		Sandbox: c.sandbox,
	}
	_, err := runtime.TerminateSandbox(ctx, req)
	return err
}

func (c *JobRunContext) makeStepRunContexts() {
	c.stepContexts = make(map[string]*StepRunContext, len(c.JobRun.Steps))
	for _, step := range c.JobRun.Steps {
		c.stepContexts[step.StepId()] = NewStepRunContext(c, step)
	}
}

func (c *JobRunContext) initializeSteps(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range c.JobRun.Steps {
		rCtx := c.stepContexts[step.StepId()]
		g.Go(func() error {
			return rCtx.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (c *JobRunContext) runStage(ctx context.Context, stage Stage, fn func(executor StepRun) *Task) error {
	ids := make([]string, len(c.JobRun.Steps))
	for i, step := range c.JobRun.Steps {
		ids[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(ids) // in place reverse
	}
	for _, id := range ids {
		rCtx := c.stepContexts[id]
		if err := rCtx.RunStep(ctx, fn); err != nil {
			return err
		}
	}

	return nil
}

// Add paths to the context and remove duplicates
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path
func (c *JobRunContext) addPath(paths []string) error {
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
	for _, path := range c.paths {
		if !set.Has(path) {
			newPaths = append(newPaths, path)
			set.Insert(path)
		}
	}

	c.paths = newPaths
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L132
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-environment-variable
func (c *JobRunContext) setEnv(env map[string]string) error {
	for k, v := range env {
		if setEnvBlockList.Has(k) {
			// TODO context.AddIssue
			continue
		}
		c.env[k] = v
	}
	return nil
}

func (c *JobRunContext) toContainerConfig(ctx context.Context, container *workflows.Container) (*container.ContainerConfig, error) {
	return nil, nil
}
