package executor

import (
	"context"
	"slices"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
	"golang.org/x/sync/errgroup"
)

type compositeActionRun struct {
	action *actions.CompositeRuns

	stepRuns     []StepRun
	stepContexts map[string]*StepRunContext
}

func (ar *compositeActionRun) Initialize(ctx context.Context, rCtx *StepRunContext) (err error) {
	ar.stepContexts = make(map[string]*StepRunContext, len(ar.stepRuns))
	for _, step := range ar.stepRuns {
		ar.stepContexts[step.StepId()] = rCtx.NewChildContext(step)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, step := range ar.stepRuns {
		rChildCtx := ar.stepContexts[step.StepId()]
		g.Go(func() error {
			return rChildCtx.Initialize(ctx)
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
		Run: func(ctx context.Context, _ *StepRunContext) error {
			for _, id := range taskIds {
				rChildCtx := ar.stepContexts[id]
				if err := rChildCtx.RunStep(ctx, fn); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
