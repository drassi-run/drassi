package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
)

type javaScriptActionExecutor struct {
	action *actions.JavaScriptRuns
	repo   *repository
	rev    string
}

func (e *javaScriptActionExecutor) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	panic("Copy repo to sandbox")
}

func (e *javaScriptActionExecutor) PreTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionExecutor) MainTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionExecutor) PostTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionExecutor) Action() actions.Runs {
	return e.action
}
