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

type ScriptStepRun struct {
	BaseStepRun

	Run        workflows.Evaluable[string]
	Shell      string
	WorkingDir workflows.Evaluable[string]

	// injected values
	jobExec  JobExecutor
	sandbox  sandboxer.Sandbox
	streams  stream.Streams
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

	defaultName := fmt.Sprintf("%s", sr.Run)
	if err := sr.evaluateDisplayName(ctx, sr.exprEnv, defaultName); err != nil {
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

	shell := model.Shell(sr.defaults.Run.Shell)
	if sr.Shell != "" {
		shell = model.Shell(sr.Shell)
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
		scribe.WithMap("env", exec.ComposeEnv(false)),
	)

	script = shell.FixupScript(script)
	scriptPath := sr.computeScriptPath(shell.Extension())
	sr.expandCommand(cmd, scriptPath)

	if err = sr.transferScriptIn(ctx, script, scriptPath); err != nil {
		return nil
	}

	env := exec.ComposeEnv(true)
	paths := sr.jobExec.SystemPaths()
	return sr.sandbox.Execute(ctx, cmd, paths, env, workdir, sr.streams)
}

func (sr *ScriptStepRun) expandCommand(cmd []string, scriptPath string) {
	scriptPath = path.Join(sr.sandbox.Layout().Temp, scriptPath)
	for i, c := range cmd {
		cmd[i] = strings.Replace(c, `{0}`, scriptPath, 1)
	}
}

func (sr *ScriptStepRun) computeScriptPath(ext string) string {
	file := sr.Uid + xstring.EnsurePrefix(ext, ".")
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
