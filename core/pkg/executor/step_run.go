package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/model"
	utilreader "drassi.run/core/pkg/util/reader"
)

type ScriptStepRun struct {
	BaseStepRun
}

func (sr *ScriptStepRun) Initialize(_ context.Context, _ *StepExecutor) error {
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

func (sr *ScriptStepRun) executeMain(ctx context.Context, exec *StepExecutor) error {
	inputs, err := sr.Inputs.Evaluate("job.step", exec.evaluationSupplier)
	if err != nil {
		return err
	}

	shell := model.Shell(sr.getShell(inputs, exec))
	workdir, ok := inputs["workdir"]
	if !ok {
		workdir = exec.job.defaults.Run.WorkingDir
	}
	script, ok := inputs["script"]
	if !ok {
		return fmt.Errorf("script not found")
	}
	script = shell.FixupScript(script)

	cmd, scriptPath, err := sr.getCommand(shell, exec)
	if err != nil {
		return err
	}

	reader, err := utilreader.FromFileEntries(ctx, &utilreader.FileEntry{
		Name:    "",
		Content: script,
	})
	if err != nil {
		return err
	}

	if err = exec.Sandbox().CopyIn(ctx, reader, scriptPath); err != nil {
		return err
	}
	return exec.Sandbox().Execute(ctx, cmd, nil, workdir, exec.Streams())
}

func (sr *ScriptStepRun) PostTask() *Task {
	return nil
}

func (sr *ScriptStepRun) getShell(inputs map[string]string, exec *StepExecutor) string {
	if shell, ok := inputs["shell"]; ok {
		return shell
	}
	return exec.job.defaults.Run.Shell
}

func (sr *ScriptStepRun) getCommand(shell model.Shell, exec *StepExecutor) ([]string, string, error) {
	scriptPath := sr.getScriptPath(exec, shell, sr.Id)
	cmds, err := shell.Command()
	if err != nil {
		return nil, "", err
	}

	c := make([]string, len(cmds))
	for i, cmd := range cmds {
		c[i] = strings.Replace(cmd, `{0}`, scriptPath, 1)
	}
	return c, scriptPath, nil
}

func (sr *ScriptStepRun) getScriptPath(exec *StepExecutor, shell model.Shell, name string) string {
	scriptName := name
	//for stepInfo := &rc.StepInfo; stepInfo.Parent != nil; stepInfo = stepInfo.Parent {
	//	scriptName = fmt.Sprintf("%s-composite-%s", stepInfo.Parent.StepId, scriptName)
	//}
	return filepath.Join(exec.Sandbox().GetTempDir(), "scripts", scriptName+shell.Extension())
}
