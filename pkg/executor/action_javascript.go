package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type javaScriptActionRunner struct {
	rCtx   *RunContext
	action *actions.JavaScriptRuns
	repo   *repository
	rev    string
}

func (e *javaScriptActionRunner) Initialize(ctx context.Context) error {
	panic("Copy repo to sandbox")
}

func (e *javaScriptActionRunner) PreTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionRunner) MainTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionRunner) PostTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *javaScriptActionRunner) Action() actions.Runs {
	return e.action
}
