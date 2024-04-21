package executor

import (
	"context"
	"fmt"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	"strings"

	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
)

type runStepRunner struct {
	step *workflows.RunStep
}

func (e *runStepRunner) Initialize(_ context.Context, _ *StepRunContext) error {
	return nil
}

func (e *runStepRunner) PreTask() *Task {
	return nil
}

func (e *runStepRunner) MainTask() *Task {
	return &Task{
		StepID:    e.step.Id,
		Stage:     StageMain,
		Condition: e.step.If,
		Run:       e.executeMain,
	}
}

func (e *runStepRunner) executeMain(ctx context.Context, rCtx *StepRunContext) error {
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

	cmd, scriptName, err := e.getCommand(shell, rCtx)
	if err != nil {
		return err
	}

	reader, err := utilreader.FromFileEntries(ctx, &utilreader.FileEntry{
		Name:    scriptName,
		Content: script,
	})
	if err != nil {
		return err
	}

	if err = rCtx.Sandbox().CopyIn(ctx, reader, scriptName); err != nil {
		return err
	}
	return rCtx.Sandbox().Execute(ctx, cmd, nil, workdir)
}

func (e *runStepRunner) PostTask() *Task {
	return nil
}

func (e *runStepRunner) Step() workflows.Step {
	return e.step
}

func (e *runStepRunner) getShell(rCtx *StepRunContext) string {
	if e.step.Shell != "" {
		return e.step.Shell
	}
	return rCtx.job.defaultRun.shell
}

func (e *runStepRunner) getWorkingDir(ctx context.Context, rCtx *StepRunContext) (string, error) {
	if e.step.WorkingDir != nil {
		return rCtx.evaluateWorkingDir(ctx, e.step.WorkingDir)
	}
	return rCtx.job.defaultRun.workDir, nil
}

func (e *runStepRunner) getCommand(shell Shell, rCtx *StepRunContext) ([]string, string, error) {
	scriptName := getScriptName(rCtx, shell, e.step.Id)
	cmds, err := shell.Command()
	if err != nil {
		return nil, "", err
	}

	c := make([]string, len(cmds))
	for i, cmd := range cmds {
		c[i] = strings.Replace(cmd, `{0}`, scriptName, 1)
	}
	return c, scriptName, nil
}

func getScriptName(rCtx *StepRunContext, shell Shell, name string) string {
	scriptName := name
	//for stepInfo := &rc.StepInfo; stepInfo.Parent != nil; stepInfo = stepInfo.Parent {
	//	scriptName = fmt.Sprintf("%s-composite-%s", stepInfo.Parent.StepId, scriptName)
	//}
	scriptName += shell.Extension()
	return fmt.Sprintf("workflow/%s", scriptName)
}
