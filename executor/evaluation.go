package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/workflows"
)

//// JobRunContext evaluations ////

func (c *JobRunContext) evaluateName(ctx context.Context, name workflows.Evaluable[string]) (string, error) {
	if name == nil {
		return "", nil
	}
	return name.Evaluate(ctx)
}

func (c *JobRunContext) evaluateIf(ctx context.Context, condition workflows.Conditional) (bool, error) {
	if condition == nil {
		return true, nil
	}
	return condition.Meet(ctx)
}

func (c *JobRunContext) evaluateStrategyMatrix(ctx context.Context, matrix workflows.Evaluable[workflows.Matrix]) (workflows.Matrix, error) {
	if matrix == nil {
		return workflows.Matrix{}, nil // TODO
	}
	return matrix.Evaluate(ctx)
}

func (c *JobRunContext) evaluateStrategyFailFast(ctx context.Context, failFast workflows.Evaluable[bool]) (bool, error) {
	if failFast == nil {
		return true, nil
	}
	return failFast.Evaluate(ctx)
}

func (c *JobRunContext) evaluateStrategyMaxParallel(ctx context.Context, maxParallel workflows.Evaluable[int64]) (int64, error) {
	if maxParallel == nil {
		return -1, nil
	}
	return maxParallel.Evaluate(ctx)
}

func (c *JobRunContext) evaluateConcurrencyGroup(ctx context.Context, group workflows.Evaluable[string]) (string, error) {
	if group == nil {
		return "", nil
	}
	return group.Evaluate(ctx)
}

func (c *JobRunContext) evaluateContinueOnError(ctx context.Context, coe workflows.Evaluable[bool]) (bool, error) {
	if coe == nil {
		return false, nil
	}
	return coe.Evaluate(ctx)
}

func (c *JobRunContext) evaluateRunsOnLabels(ctx context.Context, labels []workflows.Evaluable[string]) ([]string, error) {
	if len(labels) == 0 {
		return nil, nil
	}

	res := make([]string, 0, len(labels))
	for _, label := range labels {
		if l, err := label.Evaluate(ctx); err != nil {
			return nil, err
		} else {
			res = append(res, l)
		}
	}
	return res, nil
}

func (c *JobRunContext) evaluateEnvironment(ctx context.Context, env workflows.Evaluable[string]) (string, error) {
	if env == nil {
		return "", nil
	}
	return env.Evaluate(ctx)
}

func (c *JobRunContext) evaluateJobOutputs(ctx context.Context, outputs map[string]workflows.Evaluable[string]) (map[string]string, error) {
	if len(outputs) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(outputs))
	for k, v := range outputs {
		if value, err := v.Evaluate(ctx); err != nil {
			return nil, err
		} else {
			m[k] = value
		}
	}
	return m, nil
}

func (c *JobRunContext) evaluateEnv(ctx context.Context, env workflows.Env) (map[string]string, error) {
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

func (c *JobRunContext) evaluateWorkingDir(ctx context.Context, workDir workflows.Evaluable[string]) (string, error) {
	if workDir == nil {
		return "", nil
	}
	return workDir.Evaluate(ctx)
}

func (c *JobRunContext) evaluateTimeoutMinutes(ctx context.Context, timeout workflows.Evaluable[int64]) (int64, error) {
	if timeout == nil {
		return 360, nil
	}
	return timeout.Evaluate(ctx)
}

func (c *JobRunContext) evaluateWith(ctx context.Context, with workflows.With) (map[string]string, error) {
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

func (c *JobRunContext) evaluateJobSecrets(ctx context.Context, secrets *workflows.JobSecrets) (map[string]string, error) {
	if secrets == nil {
		return nil, nil
	}
	if secrets.Inherit {
		return make(map[string]string), nil // TODO
	}

	m := map[string]string{}
	for k, v := range secrets.Secrets {
		if value, err := v.Evaluate(ctx); err != nil {
			return nil, err
		} else {
			m[k] = value
		}
	}

	return m, nil
}

//// JobRunContext.Container evaluations ////

func (c *JobRunContext) evaluateContainerCredentials(ctx context.Context, cred workflows.Evaluable[string]) (string, error) {
	if cred == nil {
		return "", nil
	}
	return cred.Evaluate(ctx)
}

func (c *JobRunContext) evaluateContainerEnv(ctx context.Context, env workflows.Env) (map[string]string, error) {
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

//// StepRunContext evaluations ////

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
