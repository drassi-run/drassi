package executor

import (
	"context"
	"fmt"
	"slices"

	"github.com/dungdm93/drassi/core/pkg/model/actions"
	"golang.org/x/sync/errgroup"
)

type compositeActionExecutor struct {
	action *actions.CompositeRuns

	executors []StepExecutor
	contexts  map[string]*StepRunContext
}

func (e *compositeActionExecutor) Initialize(ctx context.Context, rCtx *StepRunContext) (err error) {
	steps := e.action.Steps
	e.contexts = make(map[string]*StepRunContext, len(steps))
	e.executors = make([]StepExecutor, len(steps))
	for i, step := range steps {
		e.contexts[step.Base().Id] = rCtx.NewChildContext(step)
		e.executors[i], err = NewStepExecutor(step)
		if err != nil {
			return
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, executor := range e.executors {
		r := executor
		rChildCtx := e.contexts[r.Step().Base().Id]
		g.Go(func() error {
			return r.Initialize(ctx, rChildCtx)
		})
	}
	return g.Wait()
}

func (e *compositeActionExecutor) PreTask() *Task {
	return e.createStageTask(StagePre, StepExecutor.PreTask)
}

func (e *compositeActionExecutor) MainTask() *Task {
	return e.createStageTask(StageMain, StepExecutor.MainTask)
}

func (e *compositeActionExecutor) PostTask() *Task {
	return e.createStageTask(StagePost, StepExecutor.PostTask)
}

func (e *compositeActionExecutor) Action() actions.Runs {
	return e.action
}

func (e *compositeActionExecutor) createStageTask(stage Stage, fn func(StepExecutor) *Task) *Task {
	var tasks []*Task
	for _, executor := range e.executors {
		if t := fn(executor); t != nil {
			tasks = append(tasks, t)
		}
	}
	if len(tasks) == 0 {
		return nil
	}

	return &Task{
		Stage: stage,
		Run:   e.executeTasks(stage, tasks),
	}
}

func (e *compositeActionExecutor) executeTasks(stage Stage, tasks []*Task) func(context.Context, *StepRunContext) error {
	if stage == StagePost {
		slices.Reverse(tasks) // in place reverse
	}

	return func(ctx context.Context, rCtx *StepRunContext) error {
		for _, task := range tasks {
			rChildCtx := e.contexts[task.StepID]
			if rChildCtx == nil || rChildCtx.parent != rCtx {
				return fmt.Errorf("StepRunContext for task %s need to be initialize correctly", task.StepID)
			}
			_ = rChildCtx.RunStep(ctx, task)
		}
		return nil
	}
}
