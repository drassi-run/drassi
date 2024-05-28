package executor

import (
	"context"
	"slices"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
	"golang.org/x/sync/errgroup"
)

type compositeActionExecutor struct {
	action *actions.CompositeRuns

	stepRuns     []StepRun
	stepContexts map[string]*StepRunContext
}

func (e *compositeActionExecutor) Initialize(ctx context.Context, rCtx *StepRunContext) (err error) {
	e.stepContexts = make(map[string]*StepRunContext, len(e.stepRuns))
	for _, step := range e.stepRuns {
		e.stepContexts[step.StepId()] = rCtx.NewChildContext(step)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, step := range e.stepRuns {
		rChildCtx := e.stepContexts[step.StepId()]
		g.Go(func() error {
			return rChildCtx.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (e *compositeActionExecutor) PreTask() *Task {
	return e.createStageTask(StagePre, StepRun.PreTask)
}

func (e *compositeActionExecutor) MainTask() *Task {
	return e.createStageTask(StageMain, StepRun.MainTask)
}

func (e *compositeActionExecutor) PostTask() *Task {
	return e.createStageTask(StagePost, StepRun.PostTask)
}

func (e *compositeActionExecutor) Action() actions.Runs {
	return e.action
}

func (e *compositeActionExecutor) createStageTask(stage Stage, fn func(StepRun) *Task) *Task {
	taskIds := make([]string, len(e.stepRuns))
	for i, step := range e.stepRuns {
		taskIds[i] = step.StepId()
	}
	if stage == StagePost {
		slices.Reverse(taskIds) // in place reverse
	}

	return &Task{
		Stage: stage,
		Run: func(ctx context.Context, _ *StepRunContext) error {
			for _, id := range taskIds {
				rChildCtx := e.stepContexts[id]
				if err := rChildCtx.RunStep(ctx, fn); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
