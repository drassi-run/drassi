package executor

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"slices"
	"time"

	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ExecutionEnvironment interface {
	Execute(ctx context.Context, cmd []string, env map[string]string, workdir string) error
	CopyIn(ctx context.Context, reader io.Reader, dst string) error
	CopyOut(ctx context.Context, src string) (io.Reader, error)

	RunContainer(ctx context.Context, image string, entrypoint []string, cmd []string, env map[string]string, workdir string) error
	PullImage(ctx context.Context, image string) error
	BuildImage(ctx context.Context, image string, dockerfile string, contextPath string) error

	GetWorkPath() string
	GetWorkflowPath() string
	GetActionsPath() string
}

type StepInfo struct {
	StepId string
	Parent *StepInfo
}

// alias rCtx
type RunContext struct {
	contexts.Context
	ExecutionEnvironment
	StepInfo

	Default workflows.Defaults

	paths []string
}

func (rCtx *RunContext) runStep(
	ctx context.Context,
	step *workflows.BaseStep,
	task *Task,
) error {
	stepResult := contexts.Step{
		Outputs:    map[string]string{},
		Conclusion: contexts.ActionResultSuccess,
		Outcome:    contexts.ActionResultSuccess,
	}
	if task.Stage == "main" {
		defer rCtx.setStepResult(step, stepResult)
	}

	if err := rCtx.initializeRunStep(ctx, step); err != nil {
		return err
	}

	if runStep, err := rCtx.evaluateCondition(ctx, step, task.Stage, task.Condition); err == nil {
		stepResult.Conclusion = contexts.ActionResultFailure
		stepResult.Outcome = contexts.ActionResultFailure
		return err
	} else if !runStep {
		stepResult.Conclusion = contexts.ActionResultSkipped
		stepResult.Outcome = contexts.ActionResultSkipped
		// TODO logging
		return nil
	}

	timeoutCtx, cancel, err := rCtx.evaluateTimeout(ctx, step)
	if err != nil {
		return err
	}
	defer cancel()

	err = task.Run(timeoutCtx)

	if err != nil {
		stepResult.Outcome = contexts.ActionResultFailure

		if continueOnError, parseErr := rCtx.evaluateContinueOnError(ctx, step); parseErr != nil {
			stepResult.Conclusion = contexts.ActionResultFailure
			return parseErr
		} else if continueOnError {
			stepResult.Conclusion = contexts.ActionResultSuccess
			//logger.Infof("Failed but continue next step")
			err = nil
		} else {
			stepResult.Conclusion = contexts.ActionResultFailure
		}

		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	}

	if err := rCtx.finalizeRunStep(ctx); err != nil {
		return err
	}
	return err
}

func (rCtx *RunContext) initializeRunStep(ctx context.Context, step *workflows.BaseStep) error {
	if err := rCtx.setupEnv(ctx, step); err != nil {
		return err
	}

	files := []*utilreader.FileEntry{
		{Name: "OUTPUT.txt", Mode: 0o666},
		{Name: "STATE.txt", Mode: 0o666},
		{Name: "PATH.txt", Mode: 0o666},
		{Name: "ENVS.txt", Mode: 0o666},
		{Name: "SUMMARY.md", Mode: 0o666},
	}
	if r, err := utilreader.FromFileEntries(ctx, files...); err != nil {
		return err
	} else {
		return rCtx.CopyIn(ctx, r, rCtx.GetWorkflowPath())
	}
}

func (rCtx *RunContext) finalizeRunStep(ctx context.Context) error {
	workflowPath := rCtx.GetWorkflowPath()
	if err := updateRunContext(ctx, rCtx, filepath.Join(workflowPath, "ENVS.txt"), utilreader.ParseEnvVars, rCtx.updateEnv); err != nil {
		return err
	}
	if err := updateRunContext(ctx, rCtx, filepath.Join(workflowPath, "STATE.txt"), utilreader.ParseEnvVars, rCtx.saveState); err != nil {
		return err
	}
	if err := updateRunContext(ctx, rCtx, filepath.Join(workflowPath, "OUTPUT.txt"), utilreader.ParseEnvVars, rCtx.setOutput); err != nil {
		return err
	}
	if err := updateRunContext(ctx, rCtx, filepath.Join(workflowPath, "PATH.txt"), utilreader.ReadLine, rCtx.updatePath); err != nil {
		return err
	}
	return nil
}

func (rCtx *RunContext) setStepResult(step *workflows.BaseStep, result contexts.Step) {
	rCtx.Steps[step.Id] = result
}

func (rCtx *RunContext) setupEnv(ctx context.Context, step *workflows.BaseStep) error {
	panic("implement me")
}

func (rCtx *RunContext) updateEnv(env map[string]string) error {
	panic("implement me")
}

func (rCtx *RunContext) saveState(data map[string]string) error {
	panic("implement me")
}

func (rCtx *RunContext) setOutput(data map[string]string) error {
	panic("implement me")
}

// Add paths to the context and remove duplicates
// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L107
func (rCtx *RunContext) updatePath(paths []string) error {
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
	for _, path := range rCtx.paths {
		if !set.Has(path) {
			newPaths = append(newPaths, path)
			set.Insert(path)
		}
	}

	rCtx.paths = newPaths
	return nil
}

func (rCtx *RunContext) evaluateCondition(ctx context.Context, step *workflows.BaseStep, stage Stage,
	condition workflows.Conditional) (bool, error) {
	panic("implement me")
}

func (rCtx *RunContext) evaluateTimeout(ctx context.Context, step *workflows.BaseStep) (context.Context, context.CancelFunc, error) {
	// TODO default: 360
	if step.TimeoutMinutes == nil {
		return ctx, func() {}, nil
	}

	if timeout, err := step.TimeoutMinutes.Evaluate(ctx); err != nil {
		return nil, nil, err
	} else if timeout > 0 {
		c, fn := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
		return c, fn, nil
	} else {
		return nil, nil, fmt.Errorf("require positive timeout value, got %d", timeout)
	}
}

func (rCtx *RunContext) evaluateContinueOnError(ctx context.Context, step *workflows.BaseStep) (bool, error) {
	return step.ContinueOnError.Evaluate(ctx)
}

func updateRunContext[R any](
	ctx context.Context, rCtx *RunContext,
	path string,
	parser func(reader io.Reader) (R, error),
	updater func(data R) error,
) error {
	r, err := rCtx.CopyOut(ctx, path)
	if err != nil {
		return err
	}
	data, err := parser(r)
	if err != nil {
		return err
	}
	return updater(data)
}
