package executor

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/dungdm93/drasi/pkg/model/workflows"
	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
)

type runStepExecutor struct {
	step *workflows.RunStep
}

func (e *runStepExecutor) Initialize(_ context.Context, _ *StepRunContext) error {
	return nil
}

func (e *runStepExecutor) PreTask() *Task {
	return nil
}

func (e *runStepExecutor) MainTask() *Task {
	return &Task{
		StepID:    e.step.Id,
		Stage:     StageMain,
		Condition: e.step.If,
		Run:       e.executeMain,
	}
}

func (e *runStepExecutor) executeMain(ctx context.Context, rCtx *StepRunContext) error {
	shell := Shell(e.getShell(rCtx))

	workdir, err := e.getWorkingDir(ctx, rCtx)
	if err != nil {
		return err
	}

	script, err := e.step.Run.Evaluate(ctx)
	if err != nil {
		return err
	}
	script = shell.FixupScript(script)

	cmd, scriptPath, err := e.getCommand(shell, rCtx)
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

func (e *runStepExecutor) PostTask() *Task {
	return nil
}

func (e *runStepExecutor) Step() workflows.Step {
	return e.step
}

func (e *runStepExecutor) getShell(rCtx *StepRunContext) string {
	if e.step.Shell != "" {
		return e.step.Shell
	}
	return rCtx.job.defaultRun.shell
}

func (e *runStepExecutor) getWorkingDir(ctx context.Context, rCtx *StepRunContext) (string, error) {
	if e.step.WorkingDir != nil {
		return rCtx.evaluateWorkingDir(ctx, e.step.WorkingDir)
	}
	return rCtx.job.defaultRun.workDir, nil
}

func (e *runStepExecutor) getCommand(shell Shell, rCtx *StepRunContext) ([]string, string, error) {
	scriptPath := e.getScriptPath(rCtx, shell, e.step.Id)
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

func (e *runStepExecutor) getScriptPath(rCtx *StepRunContext, shell Shell, name string) string {
	scriptName := name
	//for stepInfo := &rc.StepInfo; stepInfo.Parent != nil; stepInfo = stepInfo.Parent {
	//	scriptName = fmt.Sprintf("%s-composite-%s", stepInfo.Parent.StepId, scriptName)
	//}
	return filepath.Join(rCtx.Sandbox().GetTempDir(), "scripts", scriptName+shell.Extension())
}
