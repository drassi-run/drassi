package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type compositeActionRunner struct {
	rCtx   *RunContext
	action *actions.CompositeRuns
	runner StepsRunner
}

func (e *compositeActionRunner) Initialize(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (e *compositeActionRunner) PreTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *compositeActionRunner) MainTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *compositeActionRunner) PostTask() *Task {
	//TODO implement me
	panic("implement me")
}

func (e *compositeActionRunner) Action() actions.Runs {
	return e.action
}
