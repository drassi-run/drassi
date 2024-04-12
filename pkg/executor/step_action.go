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
	rCtx   *RunContext
	step   *workflows.BaseStep
	repo   *repository
	rev    string
	runner StepRunner // TODO ActionRunner
}

func (e *usesActionStepRunner) Initialize(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (e *usesActionStepRunner) PreTask() *Task {
	t := e.runner.PreTask()
	return fixupTask(t, e.step)
}

func (e *usesActionStepRunner) MainTask() *Task {
	t := e.runner.MainTask()
	return fixupTask(t, e.step)
}

func (e *usesActionStepRunner) PostTask() *Task {
	t := e.runner.PostTask()
	return fixupTask(t, e.step)
}

func fixupTask(t *Task, step *workflows.BaseStep) *Task {
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
