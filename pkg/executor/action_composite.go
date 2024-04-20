package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type compositeActionRunner struct {
	action *actions.CompositeRuns
	runner StepsRunner
}

func (e *compositeActionRunner) Initialize(ctx context.Context, rCtx *StepRunContext) error {
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
