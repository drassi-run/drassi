package executor

import (
	"context"
	"fmt"
	"path"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/string"
	"drassi.run/core/util/tar"
	"go.uber.org/dig"
)

type ScriptStepRun struct {
	BaseStepRun

	Run        workflows.Evaluable[string]
	Shell      string
	WorkingDir workflows.Evaluable[string]

	// injected values
	sandbox  sandboxer.Sandbox
	streams  sandboxer.Streams
	exprEnv  expression.Env
	defaults workflows.Defaults
}

func (sr *ScriptStepRun) Initialize(ctx context.Context, scope *dig.Scope) error {
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

func (sr *ScriptStepRun) executeMain(ctx context.Context, exec StepExecutor) error {
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
	scriptPath := sr.computeScriptPath(exec, shell.Extension())

	cmd, err := sr.expandCommand(shell, scriptPath)
	if err != nil {
		return err
	}

	if err = sr.transferScriptIn(ctx, script, scriptPath); err != nil {
		return nil
	}

	env := exec.ComposeEnv()
	paths := exec.JobExecutor().SystemPaths()
	return sr.sandbox.Execute(ctx, cmd, paths, env, workdir, sr.streams)
}

func (sr *ScriptStepRun) expandCommand(shell model.Shell, scriptPath string) ([]string, error) {
	command, err := shell.Command()
	if err != nil {
		return nil, err
	}

	scriptPath = path.Join(sr.sandbox.Layout().Temp, scriptPath)
	cmd := make([]string, len(command))
	for i, c := range command {
		cmd[i] = strings.Replace(c, `{0}`, scriptPath, 1)
	}
	return cmd, nil
}

func (sr *ScriptStepRun) computeScriptPath(exec StepExecutor, ext string) string {
	file := sr.Id
	for parent := exec.ParentExecutor(); parent != nil; parent = parent.ParentExecutor() {
		file = fmt.Sprintf("%s-composite-%s", StepId(parent), file)
	}
	ext = xstring.EnsurePrefix(ext, ".")
	file += ext

	return path.Join("scripts", file)
}

func (sr *ScriptStepRun) transferScriptIn(ctx context.Context, script, path string) error {
	entries := map[string]string{path: script}
	if reader, err := xtar.ContentReader(entries); err != nil {
		return err
	} else {
		return sr.sandbox.CopyIn(ctx, reader, sr.sandbox.Layout().Temp)
	}
}
