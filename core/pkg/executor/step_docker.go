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

type DockerStepRun struct {
	BaseStepRun

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

	resolvedImage string

	// injected values
	runtime runtime.Container
	exprEnv expression.Env
	repo    *repository.Repository
}

func (sr *DockerStepRun) PathTranslator() runtime.PathTranslator {
	return sr.runtime
}

func (sr *DockerStepRun) Initialize(ctx context.Context, scope *dig.Scope) error {
	if err := xdig.Populate(scope, &sr.runtime); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.exprEnv); err != nil {
		return err
	}
	if err := xdig.Populate(scope, &sr.repo); err != nil {
		return err
	}

	defer sr.addSpanAttrs(ctx)

	if err := sr.evaluateDisplayName(ctx, sr.exprEnv, sr.Image); err != nil {
		return err
	}

	if image, ok := strings.CutPrefix(sr.Image, "docker://"); ok {
		sr.resolvedImage = image
		return sr.runtime.Pull(ctx, image, nil)
	} else {
		return sr.runtime.Build(ctx)
	}
}

func (sr *DockerStepRun) PreTask() *Task {
	if sr.PreEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := sr.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePre,
		Condition: condition,
		Run:       sr.execute(StagePre),
	}
}

func (sr *DockerStepRun) MainTask() *Task {
	return &Task{
		StepId:    sr.Id,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.execute(StageMain),
	}
}

func (sr *DockerStepRun) PostTask() *Task {
	if sr.PostEntrypoint == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L430-L450
	condition := sr.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePost,
		Condition: condition,
		Run:       sr.execute(StagePost),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L22
func (sr *DockerStepRun) execute(stage Stage) TaskRun {
	return func(ctx context.Context, exec StepExecutor) error {
		sr.addSpanAttrs(ctx)

		inputs := make(map[string]string)
		if err := evaluator.Evaluate(sr.exprEnv, sr.Inputs, &inputs); err != nil {
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

func (sr *DockerStepRun) computeEntrypoint(stage Stage, inputs map[string]string) ([]string, error) {
	var ep string
	switch stage {
	case StagePre:
		ep = sr.PreEntrypoint
	case StagePost:
		ep = sr.PostEntrypoint
	case StageMain:
		ep = sr.Entrypoint
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

func (sr *DockerStepRun) computeArgs(inputs map[string]string) ([]string, error) {
	if sr.Args != nil {
		args := make([]string, 0)
		if err := evaluator.Evaluate(sr.exprEnv, sr.Args, &args); err != nil {
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

func (sr *DockerStepRun) addSpanAttrs(ctx context.Context) {
	span := trace.SpanFromContext(ctx)

	if image := sr.resolvedImage; image != "" {
		span.SetAttributes(semconv.ContainerImageName(image))
	}
}

func (sr *DockerStepRun) repr() string {
	str := "node action"
	if sr.resolvedImage != "" {
		str += fmt.Sprintf(" with image=%q", sr.resolvedImage)
	} else {
		str += fmt.Sprintf(" with Dockerfile=%q", sr.Image)
	}
	if sr.repo != nil {
		str += fmt.Sprintf(" from %q", repository.Location(sr.repo))
	}
	return str
}
