/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type NodeStepDef struct {
	Inputs  workflows.Evaluable[map[string]string]
	Outputs workflows.Evaluable[map[string]string]

	Runtime string // node20 | node22 | ...
	Main    string

	Pre   string
	PreIf workflows.Conditional

	Post   string
	PostIf workflows.Conditional
}

func (d *NodeStepDef) PrepareExecute(_ context.Context, scope *dig.Scope) (StepRun, error) {
	e := &nodeStepRun{def: d}

	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.sandbox); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.streams); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.repo); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.pathProv); err != nil {
		return nil, err
	}

	return e, nil
}

type nodeStepRun struct {
	def *NodeStepDef

	// injected values
	pathProv support.PathProvider
	exprEnv  expression.Env
	sandbox  sandboxer.Sandbox
	streams  stream.Streams
	repo     *repository.Repository
}

func (sr *nodeStepRun) Def() StepDef {
	return sr.def
}

func (sr *nodeStepRun) PreTask() *Task {
	def := sr.def
	if def.Pre == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := def.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePre,
		Condition: condition,
		Run:       sr.execute(StagePre),
	}
}

func (sr *nodeStepRun) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   sr.execute(StageMain),
	}
}

func (sr *nodeStepRun) PostTask() *Task {
	def := sr.def
	if def.Post == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := def.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePost,
		Condition: condition,
		Run:       sr.execute(StagePost),
	}
}

func (sr *nodeStepRun) execute(stage Stage) TaskRun {
	return func(ctx context.Context, exec StepExecutor) error {
		sr.addSpanAttrs(ctx, stage)

		spec := exec.StepSpec()
		scriptPath := sr.computeScriptPath(stage)
		cmd := []string{"node", scriptPath}

		inputs := make(map[string]string)
		if err := evaluator.Evaluate(sr.exprEnv, spec.Inputs, &inputs); err != nil {
			return err
		}

		scribe.GroupDetails(ctx, sr.repr(),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", exec.ComposeEnv(false)),
		)

		env := exec.ComposeEnv(true)
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		paths := sr.pathProv.SystemPaths()
		return sr.sandbox.Execute(ctx, cmd, paths, env, "", sr.streams)
	}
}

func (sr *nodeStepRun) computeScriptPath(stage Stage) string {
	var script string
	switch stage {
	case StagePre:
		script = sr.def.Pre
	case StagePost:
		script = sr.def.Post
	case StageMain:
		script = sr.def.Main
	}

	layout := sr.sandbox.Layout()
	scriptPath := filepath.Join(layout.Actions, repository.Location(sr.repo), script)
	return scriptPath
}

func (sr *nodeStepRun) addSpanAttrs(ctx context.Context, stage Stage) {
	span := trace.SpanFromContext(ctx)

	var script string
	switch stage {
	case StagePre:
		script = sr.def.Pre
	case StagePost:
		script = sr.def.Post
	case StageMain:
		script = sr.def.Main
	}

	span.SetAttributes(xotel.ActionScript(script))
}

func (sr *nodeStepRun) repr() string {
	return fmt.Sprintf("node action from %q", repository.Location(sr.repo))
}
