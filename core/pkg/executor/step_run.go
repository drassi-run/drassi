package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	utilreader "github.com/dungdm93/drassi/core/pkg/util/reader"
)

type ScriptStepRun struct {
	BaseStepRun
}

func (sr *ScriptStepRun) Initialize(_ context.Context, _ *StepRunContext) error {
	return nil
}

func (sr *ScriptStepRun) PreTask() *Task {
	return nil
}

func (sr *ScriptStepRun) MainTask() *Task {
	return &Task{
		StepID:    sr.ID,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.executeMain,
	}
}

func (sr *ScriptStepRun) executeMain(ctx context.Context, rCtx *StepRunContext) error {
	inputs, err := sr.Inputs.Evaluate("job.step", rCtx)
	if err != nil {
		return err
	}

	shell := Shell(sr.getShell(inputs, rCtx))
	workdir, ok := inputs["workdir"]
	if !ok {
		workdir = rCtx.job.defaults.Run.WorkingDir
	}
	script, ok := inputs["script"]
	if !ok {
		return fmt.Errorf("script not found")
	}
	script = shell.FixupScript(script)

	cmd, scriptPath, err := sr.getCommand(shell, rCtx)
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

	if err = rCtx.Sandbox().CopyIn(ctx, reader, scriptPath); err != nil {
		return err
	}
	return rCtx.Sandbox().Execute(ctx, cmd, nil, workdir)
}

func (sr *ScriptStepRun) PostTask() *Task {
	return nil
}

func (sr *ScriptStepRun) getShell(inputs map[string]string, rCtx *StepRunContext) string {
	if shell, ok := inputs["shell"]; ok {
		return shell
	}
	return rCtx.job.defaults.Run.Shell
}

func (sr *ScriptStepRun) getCommand(shell Shell, rCtx *StepRunContext) ([]string, string, error) {
	scriptPath := sr.getScriptPath(rCtx, shell, sr.ID)
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

func (sr *ScriptStepRun) getScriptPath(rCtx *StepRunContext, shell Shell, name string) string {
	scriptName := name
	//for stepInfo := &rc.StepInfo; stepInfo.Parent != nil; stepInfo = stepInfo.Parent {
	//	scriptName = fmt.Sprintf("%s-composite-%s", stepInfo.Parent.StepId, scriptName)
	//}
	return filepath.Join(rCtx.Sandbox().GetTempDir(), "scripts", scriptName+shell.Extension())
}
