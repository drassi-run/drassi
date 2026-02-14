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

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
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
	sandbox  sandboxer.Sandbox
	streams  stream.Streams
	exprEnv  expression.Env
	defaults workflows.Defaults
}

func (e *scriptActionExecutor) init(_ context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.sandbox); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.streams); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
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
	workdir := e.defaults.Run.WorkingDir
	if err := evaluator.Evaluate(e.exprEnv, spec.WorkingDir, &workdir); err != nil {
		return err
	}

	script := ""
	if err := evaluator.Evaluate(e.exprEnv, spec.Run, &script); err != nil {
		return err
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
	scribe.GroupDetails(ctx, "script action",
		withScript(script),
		scribe.WithPair("shell", strings.Join(cmd, " ")),
		scribe.WithPair("workdir", workdir),
		scribe.WithMap("env", e.sExec.ComposeEnv(false)),
	)

	script = shell.FixupScript(script)
	scriptPath := e.computeScriptPath(e.sExec.StepSpec(), shell.Extension())
	e.expandCommand(cmd, scriptPath)

	if err = e.transferScriptIn(ctx, script, scriptPath); err != nil {
		return nil
	}

	env := e.sExec.ComposeEnv(true)
	paths := e.sExec.JobExecutor().SystemPaths()
	return e.sandbox.Execute(ctx, cmd, paths, env, workdir, e.streams)
}

func (e *scriptActionExecutor) expandCommand(cmd []string, scriptPath string) {
	scriptPath = path.Join(e.sandbox.Layout().Temp, scriptPath)
	for i, c := range cmd {
		cmd[i] = strings.Replace(c, `{0}`, scriptPath, 1)
	}
}

func (e *scriptActionExecutor) computeScriptPath(spec *StepSpec, ext string) string {
	file := spec.Uid + xstring.EnsurePrefix(ext, ".")
	return path.Join("scripts", file)
}

func (e *scriptActionExecutor) transferScriptIn(ctx context.Context, script, path string) error {
	entries := map[string]string{path: script}
	if reader, err := xtar.ContentReader(entries); err != nil {
		return err
	} else {
		return e.sandbox.CopyIn(ctx, reader, e.sandbox.Layout().Temp)
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
