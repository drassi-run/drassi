package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
)

type ActionRun interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Action() actions.Runs
}

// ensure ActionRun implementations
var (
	_ ActionRun = (*javaScriptActionRun)(nil)
	_ ActionRun = (*dockerActionRun)(nil)
	_ ActionRun = (*compositeActionRun)(nil)
)
