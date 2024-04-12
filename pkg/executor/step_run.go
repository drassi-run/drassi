package executor

import (
	"context"
	"fmt"
	"github.com/dungdm93/drasi/pkg/model/workflows"
	"strings"

	utilreader "github.com/dungdm93/drasi/pkg/util/reader"
)

type runStepRunner struct {
	rCtx *RunContext
	step *workflows.RunStep
}

func (e *runStepRunner) Initialize(_ context.Context) error {
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

func (e *runStepRunner) executeMain(ctx context.Context) error {
	shell := Shell(e.getShell())

	workdir, err := e.getWorkingDir(ctx)
	if err != nil {
		return err
	}

	script, err := e.step.Run.Evaluate(ctx)
	if err != nil {
		return err
	}
	script = shell.FixupScript(script)

	cmd, scriptName, err := e.getCommand(shell)
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

	if err = e.rCtx.CopyIn(ctx, reader, scriptName); err != nil {
		return err
	}
	return e.rCtx.Execute(ctx, cmd, nil, workdir)
}

func (e *runStepRunner) PostTask() *Task {
	return nil
}

func (e *runStepRunner) getShell() string {
	if e.step.Shell != "" {
		return e.step.Shell
	}
	return e.rCtx.Default.Run.Shell
}

func (e *runStepRunner) getWorkingDir(ctx context.Context) (string, error) {
	if e.step.WorkingDir != nil {
		return e.step.WorkingDir.Evaluate(ctx)
	}
	if e.rCtx.Default.Run.WorkingDir != nil {
		return e.rCtx.Default.Run.WorkingDir.Evaluate(ctx)
	}
	return "", nil
}

func (e *runStepRunner) getCommand(shell Shell) ([]string, string, error) {
	scriptName := getScriptName(e.rCtx, shell, e.step.Id)
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

func getScriptName(rc *RunContext, shell Shell, name string) string {
	scriptName := name
	for stepInfo := &rc.StepInfo; stepInfo.Parent != nil; stepInfo = stepInfo.Parent {
		scriptName = fmt.Sprintf("%s-composite-%s", stepInfo.Parent.StepId, scriptName)
	}
	scriptName += shell.Extension()
	return fmt.Sprintf("workflow/%s", scriptName)
}
