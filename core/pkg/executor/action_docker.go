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

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/executor/runtime"
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/model/workflows"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/store/repository"
	"drassi.run/core/util/dig"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type DockerActionSpec struct {
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

func (spec *DockerActionSpec) CreateExecutor(ctx context.Context, scope *dig.Scope) (ActionExecutor, error) {
	e := &dockerActionExecutor{spec: spec}
	if err := e.init(ctx, scope); err != nil {
		return nil, err
	}
	return e, nil
}

type dockerActionExecutor struct {
	spec *DockerActionSpec

	resolvedImage string

	// injected values
	runtime runtime.Container
	exprEnv expression.Env
	repo    *repository.Repository
}

func (e *dockerActionExecutor) init(ctx context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &e.runtime); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &e.repo); err != nil {
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

func (e *dockerActionExecutor) PathTranslator() runtime.PathTranslator {
	return e.runtime
}

func (e *dockerActionExecutor) PreTask() *Task {
	if e.spec.PreEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := e.spec.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePre,
		Condition: condition,
		Run:       e.execute(StagePre),
	}
}

func (e *dockerActionExecutor) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   e.execute(StageMain),
	}
}

func (e *dockerActionExecutor) PostTask() *Task {
	if e.spec.PostEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := e.spec.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePost,
		Condition: condition,
		Run:       e.execute(StagePost),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L22
func (e *dockerActionExecutor) execute(stage Stage) TaskRun {
	return func(ctx context.Context, exec StepExecutor) error {
		e.addSpanAttrs(ctx)

		spec := exec.StepSpec()
		inputs := make(map[string]string)
		if err := evaluator.Evaluate(e.exprEnv, spec.Inputs, &inputs); err != nil {
			return err
		}

		entrypoint, err := e.computeEntrypoint(stage, inputs)
		if err != nil {
			return err
		}

		args, err := e.computeArgs(inputs)
		if err != nil {
			return err
		}

		scribe.GroupDetails(ctx, e.repr(),
			scribe.WithList("entrypoint", entrypoint),
			scribe.WithList("args", args),
			scribe.WithMap("with", inputs),
			scribe.WithMap("env", exec.ComposeEnv(false)),
		)

		env := exec.ComposeEnv(true)
		for k, v := range inputs {
			k = strings.ToUpper(k)
			env["INPUT_"+k] = v
		}

		return e.runtime.Run(ctx, e.resolvedImage, entrypoint, args, env)
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

func (e *dockerActionExecutor) computeArgs(inputs map[string]string) ([]string, error) {
	if e.spec.Args != nil {
		args := make([]string, 0)
		if err := evaluator.Evaluate(e.exprEnv, e.spec.Args, &args); err != nil {
			return nil, err
		}
		return args, nil
	}

	if args, ok := inputs["args"]; ok {
		return []string{args}, nil
	} else {
		return nil, nil
	}
}

func (e *dockerActionExecutor) addSpanAttrs(ctx context.Context) {
	span := trace.SpanFromContext(ctx)

	if image := e.resolvedImage; image != "" {
		span.SetAttributes(semconv.ContainerImageName(image))
	}
}

func (e *dockerActionExecutor) repr() string {
	str := "node action"
	if e.resolvedImage != "" {
		str += fmt.Sprintf(" with image=%q", e.resolvedImage)
	} else {
		str += fmt.Sprintf(" with Dockerfile=%q", e.spec.Image)
	}
	if e.repo != nil {
		str += fmt.Sprintf(" from %q", repository.Location(e.repo))
	}
	return str
}
