package executor

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"time"

	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/sandboxer"
	utilreader "drassi.run/core/pkg/util/reader"
)

type StepExecutor interface {
	JobExecutor() JobExecutor
	NewChildExecutor(stepRun StepRun) StepExecutor
	ChildExecutor(id string) StepExecutor
	ParentExecutor() StepExecutor
	RootExecutor() StepExecutor

	StepId() string
	Streams() *sandboxer.Streams
	Sandbox() sandboxer.Sandbox
	Dossier() *dossiers.Dossier

	Initialize(ctx context.Context) error
	RunStep(ctx context.Context, fn func(StepRun) *Task) error
	ComposeEnv() map[string]string

	CreateStepSummary() error
	SaveState(state map[string]string) error
	SetOutput(output map[string]string) error
}

type stepExecutor struct {
	job      JobExecutor
	parent   StepExecutor
	children map[string]StepExecutor
	stepRun  StepRun

	state   map[string]string
	result  *dossiers.Step
	dossier *dossiers.Dossier
}

func (e *stepExecutor) StepId() string {
	return e.stepRun.StepId()
}

func (e *stepExecutor) JobExecutor() JobExecutor {
	return e.job
}

func (e *stepExecutor) NewChildExecutor(stepRun StepRun) StepExecutor {
	cExec := &stepExecutor{
		job:      e.job,
		parent:   e,
		children: make(map[string]StepExecutor),
		stepRun:  stepRun,
		state:    make(map[string]string),
		result: &dossiers.Step{
			Outputs: make(map[string]string),
		},
	}

	e.children[stepRun.StepId()] = cExec
	return cExec
}

func (e *stepExecutor) ChildExecutor(id string) StepExecutor {
	return e.children[id]
}

func (e *stepExecutor) ParentExecutor() StepExecutor {
	return e.parent
}

func (e *stepExecutor) RootExecutor() StepExecutor {
	var exec StepExecutor = e
	for exec.ParentExecutor() != nil {
		exec = exec.ParentExecutor()
	}
	return exec
}

func (e *stepExecutor) Streams() *sandboxer.Streams {
	return e.job.Streams()
}

func (e *stepExecutor) Sandbox() sandboxer.Sandbox {
	return e.job.Sandbox()
}

func (e *stepExecutor) Dossier() *dossiers.Dossier {
	return e.dossier
}

func (e *stepExecutor) Initialize(ctx context.Context) error {
	return e.stepRun.Initialize(ctx, e)
}

func (e *stepExecutor) RunStep(ctx context.Context, fn func(StepRun) *Task) error {
	task := fn(e.stepRun)
	if task == nil {
		return nil
	}

	return e.runTask(ctx, task)
}

func (e *stepExecutor) runTask(ctx context.Context, task *Task) error {
	if e.parent == nil {
		e.job.Reporter().StartStep(e.StepId())
		defer func() {
			e.job.Reporter().EndStep(e.StepId(), e.result.Outcome)
		}()
	}

	base := e.stepRun.Base()
	e.dossier = e.job.NewSubDossier()
	e.stepRun.SetContextInfo(e.dossier)
	evalSupplier := &evaluationSupplier{dossier: e.dossier}

	if env, err := base.Env.Evaluate("job.step.env", evalSupplier); err != nil {
		return err
	} else {
		maps.Copy(e.dossier.Env, env)
	}

	if task.Condition != nil {
		if meet, err := task.Condition.Meet("job.step", evalSupplier); err != nil {
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

	timeout, err := base.TimeoutInMinutes.Evaluate("job.step.timeout-minutes", evalSupplier)
	if err != nil {
		return err
	}
	if timeout <= 0 {
		timeout = 360
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cancel()

	if err = e.job.StartStep(timeoutCtx, e); err != nil {
		return err
	}
	defer e.job.EndStep()

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

		if continueOnError, parseErr := base.ContinueOnError.Evaluate("job.step.continue-on-error", evalSupplier); parseErr != nil {
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

	if err = e.finalizeRunStep(ctx); err != nil {
		return err
	}
	return nil
}

func (e *stepExecutor) initializeRunStep(ctx context.Context) error {
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

func (e *stepExecutor) finalizeRunStep(ctx context.Context) error {
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

func (e *stepExecutor) ComposeEnv() map[string]string {
	// clone dossier.Env to avoid modifying
	m := maps.Clone(e.dossier.Env)

	// NOTE:
	// * INPUT_* env will be set in the step task
	// * Other default envs are set when sandbox is created

	// set STATE_* env
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
	for k, v := range e.state {
		k = "STATE_" + k
		m[k] = v
	}

	// set GITHUB_ACTION_* env
	gh := e.dossier.Github
	m["GITHUB_ACTION"] = e.stepRun.StepId()
	m["GITHUB_ACTION_REF"] = gh.ActionRef
	m["GITHUB_ACTION_REPOSITORY"] = gh.ActionRepository

	m["PATH"] = e.job.ComposePath()

	return m
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L186
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
func (e *stepExecutor) CreateStepSummary() error {
	panic("implement me")
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L260
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#sending-values-to-the-pre-and-post-actions
func (e *stepExecutor) SaveState(state map[string]string) error {
	if e.parent != nil {
		return e.RootExecutor().SaveState(state)
	}
	maps.Copy(e.state, state)
	return nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L293
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter
func (e *stepExecutor) SetOutput(output map[string]string) error {
	maps.Copy(e.result.Outputs, output)
	return nil
}

func updateRunContext[R any](
	ctx context.Context, exe StepExecutor,
	path string,
	parser func(reader io.Reader) (R, error),
	updater func(data R) error,
) error {
	r, err := exe.Sandbox().CopyOut(ctx, path)
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
