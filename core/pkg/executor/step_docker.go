package executor

import (
	"context"
	"fmt"
	"maps"
	"strings"

	"drassi.run/core/pkg/executor/evaluator"
	"drassi.run/core/pkg/model/workflows"
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
}

func (sr *DockerStepRun) Initialize(ctx context.Context, exec StepExecutor) error {
	if image, ok := strings.CutPrefix(sr.Image, "docker://"); ok {
		sr.resolvedImage = image
		return exec.Sandbox().PullImage(ctx, image)
	}
	// TODO: build image based on Dockerfile
	return nil
}

func (sr *DockerStepRun) PreTask() *Task {
	if sr.PreEntrypoint == "" {
		return nil
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePre,
		Condition: sr.PreIf,
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
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePost,
		Condition: sr.PostIf,
		Run:       sr.execute(StagePost),
	}
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/Handlers/ContainerActionHandler.cs#L22
func (sr *DockerStepRun) execute(stage Stage) TaskRun {
	return func(ctx context.Context, exec StepExecutor) error {
		inputs := make(map[string]string)
		if err := evaluator.Evaluate(exec.ExpressionEnv(), sr.Inputs, &inputs); err != nil {
			return err
		}

		entrypoint, err := sr.computeEntrypoint(stage, inputs)
		if err != nil {
			return err
		}

		args, err := sr.computeArgs(inputs, exec)
		if err != nil {
			return err
		}

		env := make(map[string]string)
		if err = evaluator.Evaluate(exec.ExpressionEnv(), sr.Env, &env); err != nil {
			return err
		}
		maps.Copy(env, exec.ComposeEnv())

		return exec.Sandbox().RunContainer(ctx, sr.resolvedImage, entrypoint, args, env, "")
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

func (sr *DockerStepRun) computeArgs(inputs map[string]string, exec StepExecutor) ([]string, error) {
	if sr.Args != nil {
		args := make([]string, 0)
		if err := evaluator.Evaluate(exec.ExpressionEnv(), sr.Args, &args); err != nil {
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
