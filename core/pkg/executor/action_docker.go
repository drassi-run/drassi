/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package executor

import (
	"context"
	"fmt"
	"strings"

	"drassi.run/core/pkg/executor/runtime"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/util/dig"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type DockerActionSpec struct {
	Repo    *repository.Repository
	Inputs  workflows.Evaluable[map[string]string]
	Outputs workflows.Evaluable[map[string]string]
	Env     workflows.Evaluable[map[string]string]

	// Should be either:
	// * docker://image[:tag]
	// * [/path/to/]Dockerfile
	Image string

	// Entrypoint for the main stage, if empty, `with.entrypoint` could be used
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
	Entrypoint string
	// Arguments for all stages, if empty, `with.args` could be used
	// https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
	Args workflows.Evaluable[[]string]

	PreEntrypoint string
	PreIf         workflows.Conditional

	PostEntrypoint string
	PostIf         workflows.Conditional
}

func (spec *DockerActionSpec) CreateExecutor(
	ctx context.Context, scope *dig.Scope, exec StepExecutor,
) (ActionExecutor, error) {
	e := &dockerActionExecutor{spec: spec, sExec: exec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type dockerActionExecutor struct {
	spec  *DockerActionSpec
	sExec StepExecutor

	resolvedImage string

	// injected values
	runtime runtime.Container
}

func (e *dockerActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.runtime); err != nil {
		return err
	}
	defer e.addSpanAttrs(ctx)

	//if err := d.evaluateDisplayName(ctx, d.exprEnv, d.Image); err != nil {
	//	return nil, err
	//}

	if image, ok := strings.CutPrefix(e.spec.Image, "docker://"); ok {
		e.resolvedImage = image
		return e.runtime.Pull(ctx, image, nil)
	}

	return e.runtime.Build(ctx)
}

func (e *dockerActionExecutor) ActionSpec() ActionSpec {
	return e.spec
}

func (e *dockerActionExecutor) StepExecutor() StepExecutor {
	return e.sExec
}

func (e *dockerActionExecutor) Name() workflows.Evaluable[string] {
	return workflows.NewLiteralToken(e.spec.Image)
}

func (e *dockerActionExecutor) Env() workflows.Evaluable[map[string]string] {
	return e.spec.Env
}

func (e *dockerActionExecutor) Inputs() workflows.Evaluable[map[string]string] {
	return e.spec.Inputs
}

func (e *dockerActionExecutor) Outputs() workflows.Evaluable[map[string]string] {
	return e.spec.Outputs
}

func (e *dockerActionExecutor) PathTranslator() runtime.PathTranslator {
	return e.runtime
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
func (e *dockerActionExecutor) CreateTask(stage Stage) *ActionTask {
	var condition workflows.Conditional
	switch stage {
	case StagePre:
		if e.spec.PreEntrypoint == "" {
			return nil
		}
		condition = e.spec.PreIf
	case StagePost:
		if e.spec.PostEntrypoint == "" {
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

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L22
func (e *dockerActionExecutor) execute(stage Stage) ActionRun {
	return func(ctx context.Context) error {
		e.addSpanAttrs(ctx)

		inputs := e.sExec.Inputs()
		entrypoint, err := e.computeEntrypoint(stage, inputs)
		if err != nil {
			return err
		}
		args, err := e.computeArgs(e.sExec.ExprEnv(), inputs)
		if err != nil {
			return err
		}

		scribe.GroupDetails(ctx, e.repr(),
			scribe.WithList("entrypoint", entrypoint),
			scribe.WithList("args", args),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", e.sExec.Env()),
		)

		env := composeEnv(e.sExec)
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		streams := e.sExec.Streams(ctx)
		defer streams.Close()
		return e.runtime.Run(ctx, e.resolvedImage, entrypoint, args, env, streams)
	}
}

func (e *dockerActionExecutor) computeEntrypoint(stage Stage, inputs map[string]string) ([]string, error) {
	var ep string
	switch stage {
	case StagePre:
		ep = e.spec.PreEntrypoint
	case StagePost:
		ep = e.spec.PostEntrypoint
	case StageMain:
		ep = e.spec.Entrypoint
	default:
		return nil, fmt.Errorf("unknown stage: %s", stage)
	}

	if ep != "" {
		return []string{ep}, nil
	}

	if stage == StageMain {
		if ep, ok := inputs["entrypoint"]; ok {
			return []string{ep}, nil
		}
	}
	return nil, nil
}

func (e *dockerActionExecutor) computeArgs(exprEnv expression.Env, inputs map[string]string) ([]string, error) {
	if e.spec.Args != nil {
		args := make([]string, 0)
		if err := evaluator.Evaluate(exprEnv, e.spec.Args, &args); err != nil {
			return nil, fmt.Errorf("evaluate 'args': %w", err)
		}
		return args, nil
	}

	if args, ok := inputs["args"]; ok {
		return []string{args}, nil
	}

	return nil, nil
}

func (e *dockerActionExecutor) addSpanAttrs(ctx context.Context) {
	span := trace.SpanFromContext(ctx)

	if image := e.resolvedImage; image != "" {
		span.SetAttributes(semconv.ContainerImageName(image))
	}
}

func (e *dockerActionExecutor) repr() string {
	str := "docker action"
	if e.resolvedImage != "" {
		str += fmt.Sprintf(" with image=%q", e.resolvedImage)
	} else {
		str += fmt.Sprintf(" with Dockerfile=%q", e.spec.Image)
	}
	if repo := e.spec.Repo; repo != nil {
		str += fmt.Sprintf(" from %q", repository.Location(repo))
	}
	return str
}
