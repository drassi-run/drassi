package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
)

type javaScriptActionRun struct {
	action *actions.JavaScriptRuns
	repo   *Repository
	rev    string
}

func (ar *javaScriptActionRun) Initialize(ctx context.Context, exec *StepExecutor) error {
	panic("Copy repo to sandbox")
}

func (ar *javaScriptActionRun) PreTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (ar *javaScriptActionRun) MainTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (ar *javaScriptActionRun) PostTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (ar *javaScriptActionRun) Action() actions.Runs {
	return ar.action
}
