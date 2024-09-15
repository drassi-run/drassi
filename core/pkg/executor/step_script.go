package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	utilreader "drassi.run/core/pkg/util/reader"
)

type ScriptStepRun struct {
	BaseStepRun

	Run        workflows.Evaluable[string]
	Shell      string
	WorkingDir workflows.Evaluable[string]
}

func (sr *ScriptStepRun) Initialize(_ context.Context, _ StepExecutor) error {
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

func (sr *ScriptStepRun) executeMain(ctx context.Context, exec StepExecutor) error {
	shell := model.Shell(exec.JobExecutor().Defaults().Run.Shell)
	if sr.Shell != "" {
		shell = model.Shell(sr.Shell)
	}

	workdir := exec.JobExecutor().Defaults().Run.WorkingDir
	if err := evaluator.Evaluate(exec.ExpressionEnv(), sr.WorkingDir, &workdir); err != nil {
		return err
	}

	script := ""
	if err := evaluator.Evaluate(exec.ExpressionEnv(), sr.Run, &script); err != nil {
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

	if err = sr.transferScriptIn(ctx, exec, script, path); err != nil {
		return nil
	}

	env := exec.ComposeEnv()
	return exec.Sandbox().Execute(ctx, cmd, env, workdir, exec.Streams())
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

func (sr *ScriptStepRun) computeScriptPath(exec StepExecutor, ex string) string {
	path := sr.Id
	for parent := exec.ParentExecutor(); parent != nil; parent = parent.ParentExecutor() {
		path = fmt.Sprintf("%s-composite-%s", parent.StepId(), path)
	}
	if !strings.HasPrefix(ex, ".") {
		path += "."
	}
	path += ex
	return filepath.Join(exec.Sandbox().GetTempDir(), "scripts", path)
}

func (sr *ScriptStepRun) transferScriptIn(ctx context.Context, exec StepExecutor, script, path string) error {
	entry := &utilreader.FileEntry{
		Name:    "",
		Content: script,
		Mode:    0o755,
	}
	if reader, err := utilreader.FromFileEntries(ctx, entry); err != nil {
		return err
	} else {
		return exec.Sandbox().CopyIn(ctx, reader, path)
	}
}
