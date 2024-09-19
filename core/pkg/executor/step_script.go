package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/util/dig"
	utilreader "drassi.run/core/pkg/util/reader"

	"go.uber.org/dig"
)

type ScriptStepRun struct {
	BaseStepRun

	Run        workflows.Evaluable[string]
	Shell      string
	WorkingDir workflows.Evaluable[string]

	// injected values
	sandbox  sandboxer.Sandbox
	streams  *sandboxer.Streams
	exprEnv  expression.Env
	defaults *workflows.Defaults
}

func (sr *ScriptStepRun) Initialize(exec StepExecutor, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &sr.sandbox); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.streams); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.defaults); err != nil {
		return err
	}

	return nil
}

func (sr *ScriptStepRun) PreTask() *Task {
	return nil
}

func (sr *ScriptStepRun) MainTask() *Task {
	return &Task{
		StepId:    sr.Id,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.executeMain,
	}
}

func (sr *ScriptStepRun) PostTask() *Task {
	return nil
}

func (sr *ScriptStepRun) executeMain(exec StepExecutor) error {
	ctx := exec.Context()
	shell := model.Shell(sr.defaults.Run.Shell)
	if sr.Shell != "" {
		shell = model.Shell(sr.Shell)
	}

	workdir := sr.defaults.Run.WorkingDir
	if err := evaluator.Evaluate(sr.exprEnv, sr.WorkingDir, &workdir); err != nil {
		return err
	}

	script := ""
	if err := evaluator.Evaluate(sr.exprEnv, sr.Run, &script); err != nil {
		return err
	} else if script == "" {
		return fmt.Errorf("script is required")
	}

	script = shell.FixupScript(script)
	path := sr.computeScriptPath(exec, shell.Extension())

	cmd, err := sr.expandCommand(shell, path)
	if err != nil {
		return err
	}

	if err = sr.transferScriptIn(ctx, script, path); err != nil {
		return nil
	}

	env := make(map[string]string)
	exec.ComposeEnv(env)

	return sr.sandbox.Execute(ctx, cmd, env, workdir, sr.streams)
}

func (sr *ScriptStepRun) expandCommand(shell model.Shell, scriptPath string) ([]string, error) {
	command, err := shell.Command()
	if err != nil {
		return nil, err
	}

	cmd := make([]string, len(command))
	for i, c := range command {
		cmd[i] = strings.Replace(c, `{0}`, scriptPath, 1)
	}
	return cmd, nil
}

func (sr *ScriptStepRun) computeScriptPath(exec StepExecutor, ext string) string {
	path := sr.Id
	for parent := exec.ParentExecutor(); parent != nil; parent = parent.ParentExecutor() {
		path = fmt.Sprintf("%s-composite-%s", parent.StepId(), path)
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	path += ext
	return filepath.Join(sr.sandbox.GetTempDir(), "scripts", path)
}

func (sr *ScriptStepRun) transferScriptIn(ctx context.Context, script, path string) error {
	entry := &utilreader.FileEntry{
		Name:    "",
		Content: script,
		Mode:    0o755,
	}
	if reader, err := utilreader.FromFileEntries(ctx, entry); err != nil {
		return err
	} else {
		return sr.sandbox.CopyIn(ctx, reader, path)
	}
}
