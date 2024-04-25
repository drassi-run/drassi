package executor

import (
	"context"
	"github.com/dungdm93/drasi/pkg/model/workflows"
)

// Example:
// + `uses: docker://alpine:3.8`
// + `uses: docker://gcr.io/cloud-builders/gradle`
type usesDockerStepExecutor struct {
	step  *workflows.UsesStep
	image string
}

func (e *usesDockerStepExecutor) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	return rCtx.Sandbox().PullImage(ctx, e.image)
}

func (e *usesDockerStepExecutor) PreTask() *Task {
	return nil
}

func (e *usesDockerStepExecutor) MainTask() *Task {
	return &Task{
		StepID:    e.step.Id,
		Stage:     StageMain,
		Condition: e.step.If,
		Run:       e.executeMain,
	}
}

func (e *usesDockerStepExecutor) executeMain(ctx context.Context, rCtx *StepRunContext) error {
	return rCtx.Sandbox().RunContainer(ctx, e.image, nil, nil, nil, "")
}

func (e *usesDockerStepExecutor) PostTask() *Task {
	return nil
}

func (e *usesDockerStepExecutor) Step() workflows.Step {
	return e.step
}
