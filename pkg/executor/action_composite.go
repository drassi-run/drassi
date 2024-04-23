package executor

import (
	"context"
	"fmt"
	"slices"

	"github.com/dungdm93/drasi/pkg/model/actions"
	"golang.org/x/sync/errgroup"
)

type compositeActionRunner struct {
	action *actions.CompositeRuns

	runners  []StepRunner
	contexts map[string]*StepRunContext
}

func (e *compositeActionRunner) Initialize(ctx context.Context, rCtx *StepRunContext) (err error) {
	steps := e.action.Steps
	e.contexts = make(map[string]*StepRunContext, len(steps))
	e.runners = make([]StepRunner, len(steps))
	for i, step := range steps {
		e.contexts[step.Base().Id] = rCtx.NewChildContext(step)
		e.runners[i], err = NewStepRunner(step)
		if err != nil {
			return
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, runner := range e.runners {
		r := runner
		rChildCtx := e.contexts[r.Step().Base().Id]
		g.Go(func() error {
			return r.Initialize(ctx, rChildCtx)
		})
	}
	return g.Wait()
}

func (e *compositeActionRunner) PreTask() *Task {
	return e.createStageTask(StagePre, func(runner StepRunner) *Task {
		return runner.PreTask()
	})
}

func (e *compositeActionRunner) MainTask() *Task {
	return e.createStageTask(StageMain, func(runner StepRunner) *Task {
		return runner.MainTask()
	})
}

func (e *compositeActionRunner) PostTask() *Task {
	return e.createStageTask(StagePost, func(runner StepRunner) *Task {
		return runner.PostTask()
	})
}

func (e *compositeActionRunner) Action() actions.Runs {
	return e.action
}

func (e *compositeActionRunner) createStageTask(stage Stage, fn func(StepRunner) *Task) *Task {
	var tasks []*Task
	for _, runner := range e.runners {
		if t := fn(runner); t != nil {
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

func (e *compositeActionRunner) executeTasks(stage Stage, tasks []*Task) func(context.Context, *StepRunContext) error {
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
