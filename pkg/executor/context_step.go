package executor

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	"github.com/dungdm93/drasi/pkg/sandboxer"
	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
)

type StepRunContext struct {
	job    *JobRunContext
	parent *StepRunContext
	step   workflows.Step

	envOverride map[string]string
	input       map[string]string

	result *contexts.Step
	state  map[string]string
}

func NewStepRunContext(jobContext *JobRunContext, step workflows.Step) *StepRunContext {
	return &StepRunContext{
		job:         jobContext,
		parent:      nil,
		step:        step,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &contexts.Step{},
	}
}

func (c *StepRunContext) NewChildContext(step workflows.Step) *StepRunContext {
	return &StepRunContext{
		job:         c.job,
		parent:      c,
		step:        step,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &contexts.Step{},
	}
}

func (c *StepRunContext) Sandbox() sandboxer.Sandbox {
	return c.job.Sandbox()
}

func (c *StepRunContext) RootContext() *StepRunContext {
	r := c
	for r.parent != nil {
		r = r.parent
	}
	return r
}

func (c *StepRunContext) RunStep(ctx context.Context, task *Task) error {
	base := c.step.Base()
	if task.StepID != base.Id {
		return fmt.Errorf("task step ID does not match step ID %s", task.StepID)
	}

	if err := c.setupEnv(ctx, c.step); err != nil {
		return err
	}

	if meet, err := c.evaluateIf(ctx, task.Condition); err != nil {
		c.result.Conclusion = contexts.ActionResultFailure
		c.result.Outcome = contexts.ActionResultFailure
		return err
	} else if !meet {
		c.result.Conclusion = contexts.ActionResultSkipped
		c.result.Outcome = contexts.ActionResultSkipped
		// TODO logging
		return nil
	}

	if err := c.initializeRunStep(ctx); err != nil {
		return err
	}

	timeout, err := c.evaluateTimeoutMinutes(ctx, base.TimeoutMinutes)
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()

	ch := make(chan error)
	go func() {
		ch <- task.Run(timeoutCtx, c)
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}

	if err != nil {
		c.result.Outcome = contexts.ActionResultFailure

		if continueOnError, parseErr := c.evaluateContinueOnError(ctx, base.ContinueOnError); parseErr != nil {
			c.result.Conclusion = contexts.ActionResultFailure
			return parseErr
		} else if continueOnError {
			c.result.Conclusion = contexts.ActionResultSuccess
			//logger.Infof("Failed but continue next step")
			err = nil
		} else {
			c.result.Conclusion = contexts.ActionResultFailure
		}

		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	} else {
		c.result.Conclusion = contexts.ActionResultSuccess
		c.result.Outcome = contexts.ActionResultSuccess
	}

	if err := c.finalizeRunStep(ctx); err != nil {
		return err
	}
	return nil
}

func (c *StepRunContext) initializeRunStep(ctx context.Context) error {
	files := []*utilreader.FileEntry{
		{Name: "GITHUB_OUTPUT", Mode: 0o666},
		{Name: "GITHUB_STATE", Mode: 0o666},
		{Name: "GITHUB_PATH", Mode: 0o666},
		{Name: "GITHUB_ENV", Mode: 0o666},
		{Name: "GITHUB_STEP_SUMMARY", Mode: 0o666},
	}
	if r, err := utilreader.FromFileEntries(ctx, files...); err != nil {
		return err
	} else {
		return c.Sandbox().CopyIn(ctx, r, c.Sandbox().GetWorkflowPath())
	}
}

func (c *StepRunContext) finalizeRunStep(ctx context.Context) error {
	workflowPath := c.Sandbox().GetWorkflowPath()

	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_OUTPUT"), utilreader.ParseEnvVars, c.setOutput); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_STATE"), utilreader.ParseEnvVars, c.saveState); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_PATH"), utilreader.ReadLine, c.job.addPath); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_ENV"), utilreader.ParseEnvVars, c.job.setEnv); err != nil {
		return err
	}
	// TODO update GITHUB_STEP_SUMMARY
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (c *StepRunContext) createStepSummary() error {
	panic("implement me")
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (c *StepRunContext) saveState(state map[string]string) error {
	// TODO if composite step -> return c.root.saveState(state)
	c.state = state
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (c *StepRunContext) setOutput(output map[string]string) error {
	c.result.Outputs = output
	return nil
}

func (c *StepRunContext) setupEnv(ctx context.Context, step workflows.Step) error {
	// TODO "implement me"
	return nil
}

func updateRunContext[R any](
	ctx context.Context, c *StepRunContext,
	path string,
	parser func(reader io.Reader) (R, error),
	updater func(data R) error,
) error {
	r, err := c.Sandbox().CopyOut(ctx, path)
	if err != nil {
		return err
	}
	data, err := parser(r)
	if err != nil {
		return err
	}
	return updater(data)
}
