package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/workflows"
)

// Example:
// + Using a public action: `uses: actions/aws@v2.0.1`
// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
// + Using a local action: `uses: ./.github/actions/hello-world-action`
type usesActionStepRunner struct {
	step   *workflows.UsesStep
	repo   *repository
	rev    string
	action ActionRunner
}

func (e *usesActionStepRunner) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	//TODO implement me
	panic("implement me")
}

func (e *usesActionStepRunner) PreTask() *Task {
	t := e.action.PreTask()
	return fixupTask(t, e.step)
}

func (e *usesActionStepRunner) MainTask() *Task {
	t := e.action.MainTask()
	return fixupTask(t, e.step)
}

func (e *usesActionStepRunner) PostTask() *Task {
	t := e.action.PostTask()
	return fixupTask(t, e.step)
}

func (e *usesActionStepRunner) Step() workflows.Step {
	return e.step
}

func fixupTask(t *Task, step *workflows.UsesStep) *Task {
	if t == nil {
		return nil
	}
	if step.If != nil {
		if t.Condition != nil {
			t.Condition = workflows.NewConditionalAnd(step.If, t.Condition)
		} else {
			t.Condition = step.If
		}
	}
	t.StepID = step.Id
	return t
}
