package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/dossiers"
	utilreader "drassi.run/core/pkg/util/reader"
)

type ScriptStepRun struct {
	BaseStepRun
}

func (sr *ScriptStepRun) SetContextInfo(dossier *dossiers.Dossier) {
	gh := dossier.Github

	gh.Action = sr.Id
	gh.ActionRepository = ""
	gh.ActionRef = ""
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

func (sr *ScriptStepRun) executeMain(ctx context.Context, exec StepExecutor) error {
	evalSupplier := &evaluationSupplier{dossier: exec.Dossier()}
	inputs, err := sr.Inputs.Evaluate("job.step", evalSupplier)
	if err != nil {
		return err
	}

	shell := sr.getShell(exec, inputs)
	workdir := sr.getWorkdir(exec, inputs)

	script, ok := inputs["script"]
	if !ok {
		return fmt.Errorf("script not found")
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

func (sr *ScriptStepRun) PostTask() *Task {
	return nil
}

func (sr *ScriptStepRun) getShell(exec StepExecutor, inputs map[string]string) model.Shell {
	if shell, ok := inputs["shell"]; ok {
		return model.Shell(shell)
	}
	return model.Shell(exec.JobExecutor().Defaults().Run.Shell)
}

func (sr *ScriptStepRun) getWorkdir(exec StepExecutor, inputs map[string]string) string {
	if workdir, ok := inputs["workdir"]; ok {
		return workdir
	}
	return exec.JobExecutor().Defaults().Run.WorkingDir
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
