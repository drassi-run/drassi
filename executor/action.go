package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type ActionRunner interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Action() actions.Runs
}

// ensure ActionRunner implementations
var (
	_ ActionRunner = (*javaScriptActionRunner)(nil)
	_ ActionRunner = (*dockerActionRunner)(nil)
	_ ActionRunner = (*compositeActionRunner)(nil)
)
