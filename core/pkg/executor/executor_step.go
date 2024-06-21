package executor

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	utilreader "drassi.run/core/pkg/util/reader"
)

type StepExecutor struct {
	job      *JobExecutor
	parent   *StepExecutor
	children map[string]*StepExecutor
	stepRun  StepRun

	evaluationSupplier workflows.EvaluationSupplier
	envOverride        map[string]string
	input              map[string]string

	result *dossiers.Step
	state  map[string]string
}

func (e *StepExecutor) StepId() string {
	return e.stepRun.StepId()
}

func (e *StepExecutor) NewChildExecutor(stepRun StepRun) *StepExecutor {
	cExec := &StepExecutor{
		job:         e.job,
		parent:      e,
		children:    make(map[string]*StepExecutor),
		stepRun:     stepRun,
		envOverride: make(map[string]string),
		input:       make(map[string]string),
		result:      &dossiers.Step{},
	}

	e.children[stepRun.StepId()] = cExec
	return cExec
}

func (e *StepExecutor) JobExecutor() *JobExecutor {
	return e.job
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

func (e *StepExecutor) Streams() *sandboxer.Streams {
	return e.job.Streams()
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
	if e.parent == nil {
		e.job.Reporter.StartStep(e.StepId())
		defer func() {
			e.job.Reporter.EndStep(e.StepId(), e.result.Outcome)
		}()
	}

	base := e.stepRun.Base()
	e.evaluationSupplier = &evaluationSupplier{dossier: e.job.NewSubDossier()}

	if err := e.setupEnv(ctx, e.stepRun); err != nil {
		return err
	}

	if task.Condition != nil {
		if meet, err := task.Condition.Meet("job.step", e.evaluationSupplier); err != nil {
			e.result.Conclusion = dossiers.ResultFailure
			e.result.Outcome = dossiers.ResultFailure
			return err
		} else if !meet {
			e.result.Conclusion = dossiers.ResultSkipped
			e.result.Outcome = dossiers.ResultSkipped
			// TODO logging
			return nil
		}
	}

	if err := e.initializeRunStep(ctx); err != nil {
		return err
	}

	timeout, err := base.TimeoutInMinutes.Evaluate("job.step.timeout-minutes", e.evaluationSupplier)
	if err != nil {
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()

	if err = e.job.consoleCmdHandler.StartStep(timeoutCtx, e); err != nil {
		return err
	}
	defer e.job.consoleCmdHandler.EndStep()

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
		e.result.Outcome = dossiers.ResultFailure

		if continueOnError, parseErr := base.ContinueOnError.Evaluate("job.step.continue-on-error", e.evaluationSupplier); parseErr != nil {
			e.result.Conclusion = dossiers.ResultFailure
			return parseErr
		} else if continueOnError {
			e.result.Conclusion = dossiers.ResultSuccess
			//logger.Infof("Failed but continue next step")
			err = nil
		} else {
			e.result.Conclusion = dossiers.ResultFailure
		}

		//logger.WithField("stepResult", stepResult.Outcome).Errorf("  \u274C  Failure - %s %s", stage, stepString)
	} else {
		e.result.Conclusion = dossiers.ResultSuccess
		e.result.Outcome = dossiers.ResultSuccess
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

	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_OUTPUT"), utilreader.ParseEnvVars, e.SetOutput); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_STATE"), utilreader.ParseEnvVars, e.SaveState); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_PATH"), utilreader.ReadLine, e.job.AddPath); err != nil {
		return err
	}
	if err := updateRunContext(ctx, e, filepath.Join(fileCommandsDir, "GITHUB_ENV"), utilreader.ParseEnvVars, e.job.SetEnv); err != nil {
		return err
	}
	// TODO update GITHUB_STEP_SUMMARY
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (e *StepExecutor) CreateStepSummary() error {
	panic("implement me")
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *StepExecutor) SaveState(state map[string]string) error {
	// TODO if composite step -> return e.root.SaveState(state)
	e.state = state
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *StepExecutor) SetOutput(output map[string]string) error {
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
