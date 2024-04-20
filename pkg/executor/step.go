package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/workflows"
)

type StepRunner interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Step() workflows.Step
}

// ensure StepRunner implementations
var (
	_ StepRunner = (*runStepRunner)(nil)
	_ StepRunner = (*usesDockerStepRunner)(nil)
	_ StepRunner = (*usesActionStepRunner)(nil)
)
