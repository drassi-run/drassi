package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
)

type ActionExecutor interface {
	Initialize(ctx context.Context, rCtx *StepRunContext) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Action() actions.Runs
}

// ensure ActionExecutor implementations
var (
	_ ActionExecutor = (*javaScriptActionExecutor)(nil)
	_ ActionExecutor = (*dockerActionExecutor)(nil)
	_ ActionExecutor = (*compositeActionExecutor)(nil)
)
