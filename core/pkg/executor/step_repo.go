package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/workflows"
)

// Example:
// + Using a public action: `uses: actions/aws@v2.0.1`
// + Using a public action in a subdirectory: `uses: actions/aws/ec2@main`
// + Using a local action: `uses: ./.github/actions/hello-world-action`
type RepositoryStepRun struct {
	BaseStepRun
	Repo Repository

	rev    string
	action ActionExecutor
}

func (sr *RepositoryStepRun) Initialize(ctx context.Context, rCtx *StepRunContext) error {
	//TODO implement me
	panic("implement me")
}

func (sr *RepositoryStepRun) PreTask() *Task {
	t := sr.action.PreTask()
	return sr.fixupTask(t)
}

func (sr *RepositoryStepRun) MainTask() *Task {
	t := sr.action.MainTask()
	return sr.fixupTask(t)
}

func (sr *RepositoryStepRun) PostTask() *Task {
	t := sr.action.PostTask()
	return sr.fixupTask(t)
}

func (sr *RepositoryStepRun) fixupTask(t *Task) *Task {
	if t == nil {
		return nil
	}
	if sr.Condition != nil {
		if t.Condition != nil {
			t.Condition = workflows.NewConditionalAnd(sr.Condition, t.Condition)
		} else {
			t.Condition = sr.Condition
		}
	}
	t.StepID = sr.ID
	return t
}
