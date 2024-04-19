package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/model/workflows"
)

type StepRunContext struct {
	*JobRunContext
	parent *StepRunContext

	envOverride map[string]string
	input       map[string]string

	result *contexts.Step
	state  map[string]string
}

func rootStepRunContext(jobContext *JobRunContext) *StepRunContext {
	return &StepRunContext{
		JobRunContext: jobContext,
		parent:        nil,
		envOverride:   make(map[string]string),
		input:         make(map[string]string),
		result:        &contexts.Step{},
	}
}

func (c *StepRunContext) newChildContext() *StepRunContext {
	return &StepRunContext{
		JobRunContext: c.JobRunContext,
		parent:        c,
		envOverride:   make(map[string]string),
		input:         make(map[string]string),
		result:        &contexts.Step{},
	}
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
