package executor

import (
	"context"
	"slices"

	"github.com/dungdm93/drasi/pkg/model/workflows"
	"github.com/dungdm93/drasi/pkg/sandboxer"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	setEnvBlockList = sets.New("NODE_OPTIONS")
)

type JobRunContext struct {
	sandbox sandboxer.Sandbox
	job     *workflows.NormalJob

	defaultRun struct {
		shell   string
		workDir string
	}

	env   map[string]string
	paths []string

	stepContexts map[string]*StepRunContext
	runners      []StepRunner
}

func (c *JobRunContext) Sandbox() sandboxer.Sandbox {
	return c.sandbox
}

func (c *JobRunContext) RunJob(ctx context.Context) error {
	if err := c.makeRunners(); err != nil {
		return err
	}

	if err := c.initializeRunners(ctx); err != nil {
		return err
	}

	if err := c.runStage(ctx, StagePre, StepRunner.PreTask); err != nil {
		return err
	}
	if err := c.runStage(ctx, StageMain, StepRunner.MainTask); err != nil {
		return err
	}
	if err := c.runStage(ctx, StagePost, StepRunner.PostTask); err != nil {
		return err
	}
	return nil
}

func (c *JobRunContext) makeRunners() error {
	c.stepContexts = make(map[string]*StepRunContext, len(c.job.Steps))
	c.runners = make([]StepRunner, len(c.job.Steps))
	var err error
	for i, step := range c.job.Steps {
		c.stepContexts[step.Base().Id] = NewStepRunContext(c, step)
		c.runners[i], err = NewStepRunner(step)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *JobRunContext) initializeRunners(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	for i, step := range c.job.Steps {
		runner := c.runners[i]
		rCtx := c.stepContexts[step.Base().Id]
		g.Go(func() error {
			return runner.Initialize(ctx, rCtx)
		})
	}
	return g.Wait()
}

func (c *JobRunContext) runStage(ctx context.Context, stage Stage, fn func(runner StepRunner) *Task) error {
	var tasks []*Task
	for _, runner := range c.runners {
		if t := fn(runner); t != nil {
			tasks = append(tasks, t)
		}
	}
	if stage == StagePost {
		slices.Reverse(tasks) // in place reverse
	}
	for _, task := range tasks {
		rCtx := c.stepContexts[task.StepID]
		if err := rCtx.RunStep(ctx, task); err != nil {
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
