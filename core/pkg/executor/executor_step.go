package executor

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/dungdm93/drassi/core/pkg/model/contexts"
	"github.com/dungdm93/drassi/core/pkg/sandboxer"
	utilreader "github.com/dungdm93/drassi/core/pkg/util/reader"
)

type StepExecutor struct {
	job      *JobExecutor
	parent   *StepExecutor
	children map[string]*StepExecutor
	stepRun  StepRun

	envOverride map[string]string
	input       map[string]string

	result *contexts.Step
	state  map[string]string
}

func NewStepExecutor(jobExecutor *JobExecutor, stepRun StepRun) *StepExecutor {
	return &StepExecutor{
		job:         jobExecutor,
		parent:      nil,
		children:    make(map[string]*StepExecutor),
		stepRun:     stepRun,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &contexts.Step{},
	}
}

func (e *StepExecutor) NewChildExecutor(stepRun StepRun) *StepExecutor {
	cExec := &StepExecutor{
		job:         e.job,
		parent:      e,
		children:    make(map[string]*StepExecutor),
		stepRun:     stepRun,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &contexts.Step{},
	}

	e.children[stepRun.StepId()] = cExec
	return cExec
}

func (e *StepExecutor) ChildExecutor(id string) *StepExecutor {
	return e.children[id]
}

func (e *StepExecutor) ParentExecutor() *StepExecutor {
	return e.parent
}

func (e *StepExecutor) RootExecutor() *StepExecutor {
	exec := e
	for exec.parent != nil {
		exec = exec.parent
	}
	return exec
}

func (e *StepExecutor) ContextData(name string) context.Context {
	return context.Background() // TODO real implementation
}

func (e *StepExecutor) Functions(name string) []string {
	return nil // TODO real implementation
}

func (e *StepExecutor) DefaultValue(name string) any {
	switch name {
	case "job.step.timeout-minutes":
		return int64(360)
	case "job.step.working-directory":
		return e.job.defaults.Run.WorkingDir
	default:
		return nil
	}
}

func (e *StepExecutor) Sandbox() sandboxer.Sandbox {
	return e.job.Sandbox()
}

func (e *StepExecutor) Initialize(ctx context.Context) error {
	return e.stepRun.Initialize(ctx, e)
}

func (e *StepExecutor) RunStep(ctx context.Context, fn func(StepRun) *Task) error {
	task := fn(e.stepRun)
	if task == nil {
		return nil
	}

	return e.runTask(ctx, task)
}

func (e *StepExecutor) runTask(ctx context.Context, task *Task) error {
	base := e.stepRun.Base()

	if err := e.setupEnv(ctx, e.stepRun); err != nil {
		return err
	}

	if task.Condition != nil {
		if meet, err := task.Condition.Meet("job.step", e); err != nil {
			e.result.Conclusion = contexts.ActionResultFailure
			e.result.Outcome = contexts.ActionResultFailure
			return err
		} else if !meet {
			e.result.Conclusion = contexts.ActionResultSkipped
			e.result.Outcome = contexts.ActionResultSkipped
			// TODO logging
			return nil
		}
	}

	if err := e.initializeRunStep(ctx); err != nil {
		return err
	}

	timeout, err := base.TimeoutInMinutes.Evaluate("job.step.timeout-minutes", e)
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()

	ch := make(chan error)
	go func() {
		ch <- task.Run(timeoutCtx, e)
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-ch:
	}

	if err != nil {
		e.result.Outcome = contexts.ActionResultFailure

		if continueOnError, parseErr := base.ContinueOnError.Evaluate("job.step.continue-on-error", e); parseErr != nil {
			e.result.Conclusion = contexts.ActionResultFailure
			return parseErr
		} else if continueOnError {
			e.result.Conclusion = contexts.ActionResultSuccess
			//logger.Infof("Failed but continue next step")
			err = nil
		} else {
			e.result.Conclusion = contexts.ActionResultFailure
		}

		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	} else {
		e.result.Conclusion = contexts.ActionResultSuccess
		e.result.Outcome = contexts.ActionResultSuccess
	}

	if err := e.finalizeRunStep(ctx); err != nil {
		return err
	}
	return nil
}

func (e *StepExecutor) initializeRunStep(ctx context.Context) error {
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
		fileCommandsDir := filepath.Join(e.Sandbox().GetTempDir(), "file_commands")
		return e.Sandbox().CopyIn(ctx, r, fileCommandsDir)
	}
}

func (e *StepExecutor) finalizeRunStep(ctx context.Context) error {
	fileCommandsDir := filepath.Join(e.Sandbox().GetTempDir(), "file_commands")

	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_OUTPUT"), utilreader.ParseEnvVars, e.setOutput); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_STATE"), utilreader.ParseEnvVars, e.saveState); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_PATH"), utilreader.ReadLine, e.job.addPath); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_ENV"), utilreader.ParseEnvVars, e.job.setEnv); err != nil {
		return err
	}
	// TODO update GITHUB_STEP_SUMMARY
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (e *StepExecutor) createStepSummary() error {
	panic("implement me")
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *StepExecutor) saveState(state map[string]string) error {
	// TODO if composite step -> return e.root.saveState(state)
	e.state = state
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *StepExecutor) setOutput(output map[string]string) error {
	e.result.Outputs = output
	return nil
}

func (e *StepExecutor) setupEnv(ctx context.Context, step StepRun) error {
	// TODO "implement me"
	return nil
}

func updateRunContext[R any](
	ctx context.Context, c *StepExecutor,
	path string,
	parser func(reader io.Reader) (R, error),
	updater func(data R) error,
) error {
	r, err := c.Sandbox().CopyOut(ctx, path)
	if err != nil {
		return err
	}
	defer r.Close()

	return utilreader.Untar(r, func(hdr *tar.Header, reader io.Reader) error {
		if hdr.Name != "" {
			return fmt.Errorf("expected read single file with empty name, got %s", hdr.Name)
		}

		if data, err := parser(reader); err != nil {
			return err
		} else {
			return updater(data)
		}
	})
}
