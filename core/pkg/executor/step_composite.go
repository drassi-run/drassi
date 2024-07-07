package executor

import (
	"context"
	"fmt"
	"slices"

	"drassi.run/core/pkg/model/dossiers"
	"golang.org/x/sync/errgroup"
)

type CompositeStepRun struct {
	BaseStepRun

	StepRuns []StepRun
}

func (sr *CompositeStepRun) SetContextInfo(dossier *dossiers.Dossier) {
}

func (sr *CompositeStepRun) Initialize(ctx context.Context, exec StepExecutor) (err error) {
	g, ctx := errgroup.WithContext(ctx)
	for _, step := range sr.StepRuns {
		cExec := exec.NewChildExecutor(step)
		g.Go(func() error {
			return cExec.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (sr *CompositeStepRun) PreTask() *Task {
	return sr.createStageTask(StagePre, StepRun.PreTask)
}

func (sr *CompositeStepRun) MainTask() *Task {
	return sr.createStageTask(StageMain, StepRun.MainTask)
}

func (sr *CompositeStepRun) PostTask() *Task {
	return sr.createStageTask(StagePost, StepRun.PostTask)
}

func (sr *CompositeStepRun) createStageTask(stage Stage, fn func(StepRun) *Task) *Task {
	taskIds := make([]string, len(sr.StepRuns))
	for i, step := range sr.StepRuns {
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
					return fmt.Errorf(`task %q has no child context`, id)
				}
				res := cExec.RunStep(ctx, fn)
				if res != nil && res.Conclusion == dossiers.ResultFailure {
					return fmt.Errorf(`step %q (%s) failed`, id, stage)
				}
			}
			return nil
		},
	}
}
