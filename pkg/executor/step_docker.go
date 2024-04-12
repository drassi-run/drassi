package executor

import (
	"context"
	"github.com/dungdm93/drasi/pkg/model/workflows"
)

// Example:
// + `uses: docker://alpine:3.8`
// + `uses: docker://gcr.io/cloud-builders/gradle`
type usesDockerStepRunner struct {
	rCtx  *RunContext
	step  *workflows.BaseStep
	image string
}

func (e *usesDockerStepRunner) Initialize(ctx context.Context) error {
	return e.rCtx.PullImage(ctx, e.image)
}

func (e *usesDockerStepRunner) PreTask() *Task {
	return nil
}

func (e *usesDockerStepRunner) MainTask() *Task {
	return &Task{
		StepID:    e.step.Id,
		Stage:     StageMain,
		Condition: e.step.If,
		Run:       e.executeMain,
	}
}

func (e *usesDockerStepRunner) executeMain(ctx context.Context) error {
	return e.rCtx.RunContainer(ctx, e.image, nil, nil, nil, "")
}

func (e *usesDockerStepRunner) PostTask() *Task {
	return nil
}
