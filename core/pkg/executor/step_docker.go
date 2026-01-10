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

type DockerStepDef struct {
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

func (d *DockerStepDef) PrepareExecute(ctx context.Context, scope *dig.Scope) (StepRun, error) {
	e := &dockerStepRun{def: d}

	if err := xdig.Populate(scope, &e.runtime); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.exprEnv); err != nil {
		return nil, err
	}
	if err := xdig.Populate(scope, &e.repo); err != nil {
		return nil, err
	}
	if err := e.Init(ctx); err != nil {
		return nil, err
	}

	return e, nil
}

type dockerStepRun struct {
	def *DockerStepDef

	resolvedImage string

	// injected values
	runtime runtime.Container
	exprEnv expression.Env
	repo    *repository.Repository
}

func (sr *dockerStepRun) Def() StepDef {
	return sr.def
}

func (sr *dockerStepRun) PathTranslator() runtime.PathTranslator {
	return sr.runtime
}

func (sr *dockerStepRun) Init(ctx context.Context) error {
	defer sr.addSpanAttrs(ctx)

	//if err := d.evaluateDisplayName(ctx, d.exprEnv, d.Image); err != nil {
	//	return nil, err
	//}

	if image, ok := strings.CutPrefix(sr.def.Image, "docker://"); ok {
		sr.resolvedImage = image
		return sr.runtime.Pull(ctx, image, nil)
	}

	return sr.runtime.Build(ctx)
}

func (sr *dockerStepRun) PreTask() *Task {
	if sr.def.PreEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := sr.def.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePre,
		Condition: condition,
		Run:       sr.execute(StagePre),
	}
}

func (sr *dockerStepRun) MainTask() *Task {
	return &Task{
		Stage: StageMain,
		Run:   sr.execute(StageMain),
	}
}

func (sr *dockerStepRun) PostTask() *Task {
	if sr.def.PostEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := sr.def.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		Stage:     StagePost,
		Condition: condition,
		Run:       sr.execute(StagePost),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L22
func (sr *dockerStepRun) execute(stage Stage) TaskRun {
	return func(ctx context.Context, exec StepExecutor) error {
		sr.addSpanAttrs(ctx)

		spec := exec.StepSpec()
		inputs := make(map[string]string)
		if err := evaluator.Evaluate(sr.exprEnv, spec.Inputs, &inputs); err != nil {
			return err
		}

		entrypoint, err := sr.computeEntrypoint(stage, inputs)
		if err != nil {
			return err
		}

		args, err := sr.computeArgs(inputs)
		if err != nil {
			return err
		}

		scribe.GroupDetails(ctx, sr.repr(),
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

		return sr.runtime.Run(ctx, sr.resolvedImage, entrypoint, args, env)
	}
}

func (sr *dockerStepRun) computeEntrypoint(stage Stage, inputs map[string]string) ([]string, error) {
	var ep string
	switch stage {
	case StagePre:
		ep = sr.def.PreEntrypoint
	case StagePost:
		ep = sr.def.PostEntrypoint
	case StageMain:
		ep = sr.def.Entrypoint
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

func (sr *dockerStepRun) computeArgs(inputs map[string]string) ([]string, error) {
	if sr.def.Args != nil {
		args := make([]string, 0)
		if err := evaluator.Evaluate(sr.exprEnv, sr.def.Args, &args); err != nil {
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

func (sr *dockerStepRun) addSpanAttrs(ctx context.Context) {
	span := trace.SpanFromContext(ctx)

	if image := sr.resolvedImage; image != "" {
		span.SetAttributes(semconv.ContainerImageName(image))
	}
}

func (sr *dockerStepRun) repr() string {
	str := "node action"
	if sr.resolvedImage != "" {
		str += fmt.Sprintf(" with image=%q", sr.resolvedImage)
	} else {
		str += fmt.Sprintf(" with Dockerfile=%q", sr.def.Image)
	}
	if sr.repo != nil {
		str += fmt.Sprintf(" from %q", repository.Location(sr.repo))
	}
	return str
}
