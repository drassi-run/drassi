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

type NodeActionSpec struct {
	Repo    *repository.Repository
	Inputs  workflows.Evaluable[map[string]string]
	Outputs workflows.Evaluable[map[string]string]

	Runtime string // node20 | node22 | ...
	Main    string

	Pre   string
	PreIf workflows.Conditional

	Post   string
	PostIf workflows.Conditional
}

func (spec *NodeActionSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope, exec StepExecutor,
) (ActionExecutor, error) {
	e := &nodeActionExecutor{spec: spec, sExec: exec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type nodeActionExecutor struct {
	spec  *NodeActionSpec
	sExec StepExecutor

	// injected values
	exprEnv expression.Env
	sandbox sandboxer.Sandbox
	streams stream.Streams
}

func (e *nodeActionExecutor) init(_ context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.sandbox); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.streams); err != nil {
		return err
	}
	return nil
}

func (e *nodeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *nodeActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
func (e *nodeActionExecutor) CreateTask(stage Stage) *ActionTask {
	var condition workflows.Conditional
	switch stage {
	case StagePre:
		if e.spec.Pre == "" {
			return nil
		}
		condition = e.spec.PreIf
	case StagePost:
		if e.spec.Post == "" {
			return nil
		}
		condition = e.spec.PostIf
	}
	if stage != StageMain && condition == "" {
		condition = "always()"
	}

	return &ActionTask{
		Run:       e.execute(stage),
		Stage:     stage,
		Executor:  e,
		Condition: condition,
	}
}

func (e *nodeActionExecutor) execute(stage Stage) ActionRun {
	return func(ctx context.Context) error {
		e.addSpanAttrs(ctx, stage)

		spec := e.sExec.StepSpec()
		scriptPath := e.computeScriptPath(stage)
		cmd := []string{"node", scriptPath}

		inputs := make(map[string]string)
		if err := evaluator.Evaluate(e.exprEnv, spec.Inputs, &inputs); err != nil {
			return err
		}

		scribe.GroupDetails(ctx, e.repr(),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", e.sExec.ComposeEnv(false)),
		)

		env := e.sExec.ComposeEnv(true)
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		paths := e.sExec.JobExecutor().SystemPaths()
		return e.sandbox.Execute(ctx, cmd, paths, env, "", e.streams)
	}
}

func (e *nodeActionExecutor) computeScriptPath(stage Stage) string {
	var script string
	switch stage {
	case StagePre:
		script = e.spec.Pre
	case StagePost:
		script = e.spec.Post
	case StageMain:
		script = e.spec.Main
	}

	layout := e.sandbox.Layout()
	scriptPath := filepath.Join(layout.Actions, repository.Location(e.spec.Repo), script)
	return scriptPath
}

func (e *nodeActionExecutor) addSpanAttrs(ctx context.Context, stage Stage) {
	span := trace.SpanFromContext(ctx)

	var script string
	switch stage {
	case StagePre:
		script = e.spec.Pre
	case StagePost:
		script = e.spec.Post
	case StageMain:
		script = e.spec.Main
	}

	span.SetAttributes(xotel.ActionScript(script))
}

func (e *nodeActionExecutor) repr() string {
	return fmt.Sprintf("node action from %q", repository.Location(e.spec.Repo))
}
