/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"bufio"
	"context"
	"fmt"
	"path"
	"strings"

	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/string"
	"drassi.run/core/util/tar"
	"go.uber.org/dig"
)

type ScriptActionSpec struct {
	Run        workflows.Evaluable[string]
	Shell      string
	WorkingDir workflows.Evaluable[string]
}

func (spec *ScriptActionSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope, exec StepExecutor,
) (ActionExecutor, error) {
	e := &scriptActionExecutor{spec: spec, sExec: exec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type scriptActionExecutor struct {
	spec  *ScriptActionSpec
	sExec StepExecutor

	// injected values
	defaults workflows.Defaults
}

func (e *scriptActionExecutor) init(_ context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.defaults); err != nil {
		return err
	}
	return nil
}

func (e *scriptActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *scriptActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *scriptActionExecutor) Name() workflows.Evaluable[string] {
	return e.spec.Run
}

func (e *scriptActionExecutor) Env() workflows.Evaluable[map[string]string] {
	return nil
}

func (e *scriptActionExecutor) Inputs() workflows.Evaluable[map[string]string] {
	return nil
}

func (e *scriptActionExecutor) Outputs() workflows.Evaluable[map[string]string] {
	return nil
}

func (e *scriptActionExecutor) CreateTask(stage Stage) *ActionTask {
	if stage != StageMain {
		return nil
	}
	return &ActionTask{
		Run:      e.executeMain,
		Stage:    stage,
		Executor: e,
	}
}

func (e *scriptActionExecutor) executeMain(ctx context.Context) error {
	spec := e.spec
	exprEnv := e.sExec.ExprEnv()
	workdir := e.defaults.Run.WorkingDir
	if err := evaluator.Evaluate(exprEnv, spec.WorkingDir, &workdir); err != nil {
		return fmt.Errorf("evaluate 'workingDir': %w", err)
	}

	script := ""
	if err := evaluator.Evaluate(exprEnv, spec.Run, &script); err != nil {
		return fmt.Errorf("evaluate 'run': %w", err)
	} else if script == "" {
		return fmt.Errorf("script is required")
	}

	shell := model.Shell(e.defaults.Run.Shell)
	if spec.Shell != "" {
		shell = model.Shell(spec.Shell)
	}
	cmd, err := shell.Command()
	if err != nil {
		return err
	}

	// log details before fixup script
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ScriptHandler.cs#L27
	scribe.GroupDetails(ctx, "Run script action",
		withScript(script),
		scribe.WithPair("shell", strings.Join(cmd, " ")),
		scribe.WithPair("workdir", workdir),
		scribe.WithMap("env", e.sExec.Env()),
	)

	sandbox := e.sExec.Sandbox()
	script = shell.FixupScript(script)
	scriptPath := e.computeScriptPath(e.sExec.StepSpec(), shell.Extension())
	e.expandCommand(sandbox.Layout(), cmd, scriptPath)

	if err = e.transferScriptIn(ctx, sandbox, script, scriptPath); err != nil {
		return nil
	}

	env := composeEnv(e.sExec)
	paths := e.sExec.JobExecutor().SystemPaths()
	streams := e.sExec.Streams(ctx, StageMain)
	defer streams.Close()
	return sandbox.Execute(ctx, cmd, paths, env, workdir, streams)
}

func (e *scriptActionExecutor) expandCommand(layout *sandboxer.Layout, cmd []string, scriptPath string) {
	scriptPath = path.Join(layout.Temp, scriptPath)
	for i, c := range cmd {
		cmd[i] = strings.Replace(c, `{0}`, scriptPath, 1)
	}
}

func (e *scriptActionExecutor) computeScriptPath(spec *StepSpec, ext string) string {
	file := spec.Uid + xstring.EnsurePrefix(ext, ".")
	return path.Join("scripts", file)
}

func (e *scriptActionExecutor) transferScriptIn(ctx context.Context, sandbox sandboxer.Sandbox, script, path string) error {
	entries := map[string]string{path: script}
	if reader, err := xtar.ContentReader(entries); err != nil {
		return err
	} else {
		dst := sandbox.Layout().Temp
		return sandbox.CopyIn(ctx, reader, dst)
	}
}

// print script line by line
func withScript(script string) func(*scribe.Scribe) {
	return func(s *scribe.Scribe) {
		script = strings.TrimRight(script, "\r\n")
		scanner := bufio.NewScanner(strings.NewReader(script))
		for scanner.Scan() {
			line := scanner.Text()
			// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ScriptHandler.cs#L57
			s.Writef("\x1b[36;1m%s\x1b[0m", line)
		}
	}
}
