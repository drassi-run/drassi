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

	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
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
	return e, nil
}

type nodeActionExecutor struct {
	spec  *NodeActionSpec
	sExec StepExecutor
}

func (e *nodeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *nodeActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *nodeActionExecutor) Name() workflows.Evaluable[string] {
	name := repository.Location(e.spec.Repo)
	return workflows.NewLiteralToken(name)
}

func (e *nodeActionExecutor) Env() workflows.Evaluable[map[string]string] {
	return nil
}

func (e *nodeActionExecutor) Inputs() workflows.Evaluable[map[string]string] {
	return e.spec.Inputs
}

func (e *nodeActionExecutor) Outputs() workflows.Evaluable[map[string]string] {
	return e.spec.Outputs
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

		sandbox := e.sExec.Sandbox()
		scriptPath := e.computeScriptPath(sandbox.Layout(), stage)
		cmd := []string{"node", scriptPath}
		inputs := e.sExec.Inputs()

		scribe.GroupDetails(ctx, e.repr(),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", e.sExec.Env()),
		)

		env := composeEnv(e.sExec)
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		paths := e.sExec.JobExecutor().SystemPaths()
		streams := e.sExec.Streams(ctx, stage)
		defer streams.Close()
		return sandbox.Execute(ctx, cmd, paths, env, "", streams)
	}
}

func (e *nodeActionExecutor) computeScriptPath(layout *sandboxer.Layout, stage Stage) string {
	var script string
	switch stage {
	case StagePre:
		script = e.spec.Pre
	case StagePost:
		script = e.spec.Post
	case StageMain:
		script = e.spec.Main
	}

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
