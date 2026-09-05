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
	"drassi.run/core/pkg/runtime"
	"drassi.run/core/pkg/sandboxer"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/git"
	"drassi.run/core/util/dig"
	"drassi.run/core/util/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type NodeActionSpec struct {
	Repo    *gitstore.RepoReference
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
	runtime runtime.Runtime
}

func (e *nodeActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	var provider runtime.Provider
	if err := xdig.Populate(scope, &provider); err != nil {
		return err
	}

	if rt, err := provider.Get(e.spec.Runtime); err != nil {
		return err
	} else {
		e.runtime = rt
	}
	return nil
}

func (e *nodeActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *nodeActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *nodeActionExecutor) Name() workflows.Evaluable[string] {
	name := gitstore.Location(e.spec.Repo)
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
	fn := func(ctx context.Context) error {
		e.addSpanAttrs(ctx, stage)

		sandbox := e.sExec.Sandbox()
		scriptPath := e.computeScriptPath(sandbox.Layout(), stage)
		inputs := e.sExec.Inputs()

		scribe.GroupDetails(ctx, "Run "+e.repr(),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", e.sExec.Env()),
		)

		env := e.sExec.ComposeEnv()
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		paths := e.sExec.JobExecutor().Path()
		streams := e.sExec.Streams(ctx, stage)
		defer streams.Close()
		return e.runtime.Run(ctx, scriptPath, paths, env, "", streams)
	}
	return runActionE(fn)
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

	scriptPath := filepath.Join(layout.Actions, gitstore.Location(e.spec.Repo), script)
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
	return fmt.Sprintf("node action from %q", gitstore.Location(e.spec.Repo))
}
