package executor

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
)

type StepRunContext struct {
	*JobRunContext
	parent *StepRunContext
	step   workflows.Step

	envOverride map[string]string
	input       map[string]string

	result *contexts.Step
	state  map[string]string
}

func rootStepRunContext(jobContext *JobRunContext) *StepRunContext {
	return &StepRunContext{
		JobRunContext: jobContext,
		parent:        nil,
		step:          nil,
		envOverride:   make(map[string]string),
		input:         make(map[string]string),
		result:        &contexts.Step{},
	}
}

func (c *StepRunContext) newChildContext(step workflows.Step) *StepRunContext {
	return &StepRunContext{
		JobRunContext: c.JobRunContext,
		parent:        c,
		step:          step,
		envOverride:   make(map[string]string),
		input:         make(map[string]string),
		result:        &contexts.Step{},
	}
}

func (c *StepRunContext) runStep(
	ctx context.Context,
	task *Task,
) error {
	base := c.step.Base()
	if task.StepID != base.Id {
		return fmt.Errorf("task step ID does not match step ID %s", task.StepID)
	}

	if err := c.setupEnv(ctx, c.step); err != nil {
		return err
	}

	if meet, err := c.evaluateIf(ctx, task.Condition); err == nil {
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
		return c.CopyIn(ctx, r, c.GetWorkflowPath())
	}
}

func (c *StepRunContext) finalizeRunStep(ctx context.Context) error {
	workflowPath := c.GetWorkflowPath()

	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_OUTPUT"), utilreader.ParseEnvVars, c.setOutput); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_STATE"), utilreader.ParseEnvVars, c.saveState); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_PATH"), utilreader.ReadLine, c.addPath); err != nil {
		return err
	}
	if err := updateRunContext(ctx, c, filepath.Join(workflowPath, "GITHUB_ENV"), utilreader.ParseEnvVars, c.setEnv); err != nil {
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
	panic("implement me")
}

func (c *StepRunContext) evaluateName(ctx context.Context, name workflows.Evaluable[string]) (string, error) {
	return name.Evaluate(ctx)
}

func (c *StepRunContext) evaluateIf(ctx context.Context, condition workflows.Conditional) (bool, error) {
	if condition == nil {
		return true, nil
	}
	return condition.Meet(ctx)
}

func (c *StepRunContext) evaluateWith(ctx context.Context, with workflows.With) (map[string]string, error) {
	if with == nil {
		return nil, nil
	}

	m := map[string]string{}
	for k, v := range with {
		if value, err := v.Evaluate(ctx); err != nil {
			return nil, err
		} else {
			m[k] = value
		}
	}
	return m, nil
}

func (c *StepRunContext) evaluateEnv(ctx context.Context, env workflows.Env) (map[string]string, error) {
	if env == nil {
		return nil, nil
	}
	m := map[string]string{}
	for k, v := range env {
		if value, err := v.Evaluate(ctx); err != nil {
			return nil, err
		} else {
			m[k] = value
		}
	}
	return m, nil
}

func (c *StepRunContext) evaluateContinueOnError(ctx context.Context, coe workflows.Evaluable[bool]) (bool, error) {
	if coe == nil {
		return false, nil
	}
	return coe.Evaluate(ctx)
}

func (c *StepRunContext) evaluateTimeoutMinutes(ctx context.Context, timeout workflows.Evaluable[int64]) (int64, error) {
	if timeout == nil {
		return 360, nil
	}
	return timeout.Evaluate(ctx)
}

func (c *StepRunContext) evaluateRun(ctx context.Context, run workflows.Evaluable[string]) (string, error) {
	if run == nil {
		return "", nil
	}
	return run.Evaluate(ctx)
}

func (c *StepRunContext) evaluateWorkingDir(ctx context.Context, workDir workflows.Evaluable[string]) (string, error) {
	if workDir == nil {
		return "", nil
	}
	return workDir.Evaluate(ctx)
}

func updateRunContext[R any](
	ctx context.Context, c *StepRunContext,
	path string,
	parser func(reader io.Reader) (R, error),
	updater func(data R) error,
) error {
	r, err := c.CopyOut(ctx, path)
	if err != nil {
		return err
	}
	data, err := parser(r)
	if err != nil {
		return err
	}
	return updater(data)
}
