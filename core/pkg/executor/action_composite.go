package executor

import (
	"context"
	"fmt"
	"slices"

	"drassi.run/core/pkg/model/actions"
	"golang.org/x/sync/errgroup"
)

type compositeActionRun struct {
	action *actions.CompositeRuns

	stepRuns []StepRun
}

func (ar *compositeActionRun) Initialize(ctx context.Context, exec StepExecutor) (err error) {
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range ar.stepRuns {
		cExec := exec.NewChildExecutor(step)
		g.Go(func() error {
			return cExec.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (ar *compositeActionRun) PreTask() *Task {
	return ar.createStageTask(StagePre, StepRun.PreTask)
}

func (ar *compositeActionRun) MainTask() *Task {
	return ar.createStageTask(StageMain, StepRun.MainTask)
}

func (ar *compositeActionRun) PostTask() *Task {
	return ar.createStageTask(StagePost, StepRun.PostTask)
}

func (ar *compositeActionRun) Action() actions.Runs {
	return ar.action
}

func (ar *compositeActionRun) createStageTask(stage Stage, fn func(StepRun) *Task) *Task {
	taskIds := make([]string, len(ar.stepRuns))
	for i, step := range ar.stepRuns {
		taskIds[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(taskIds) // in place reverse
	}

	return &Task{
		Stage: stage,
		Run: func(ctx context.Context, exec StepExecutor) error {
			for _, id := range taskIds {
				cExec := exec.ChildExecutor(id)
				if cExec == nil {
					return fmt.Errorf(`task "%s" has no child context`, id)
				}
				if err := cExec.RunStep(ctx, fn); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
