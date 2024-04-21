package executor

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/dungdm93/drasi/pkg/model/workflows"
	"golang.org/x/sync/errgroup"
)

type StepsRunner struct {
	steps    []workflows.Step
	runners  []StepRunner
	contexts map[string]*StepRunContext
}

func (e *StepsRunner) Initialize(ctx context.Context, rCtx *StepRunContext) (err error) {
	e.contexts = make(map[string]*StepRunContext, len(e.steps))
	e.runners = make([]StepRunner, len(e.steps))
	for i, step := range e.steps {
		e.contexts[step.Base().Id] = rCtx.NewChildContext(step)
		e.runners[i], err = e.createStepRunner(step)
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

func (e *StepsRunner) createStepRunner(step workflows.Step) (StepRunner, error) {
	switch s := step.(type) {
	case *workflows.RunStep:
		r := &runStepRunner{
			step: s,
		}
		return r, nil
	case *workflows.UsesStep:
		if image, ok := strings.CutPrefix(s.Uses, "docker://"); ok {
			r := &usesDockerStepRunner{
				step:  s,
				image: image,
			}
			return r, nil
		} else {
			repo, err := parseRepository(s.Uses)
			if err != nil {
				return nil, err
			}
			r := &usesActionStepRunner{
				step: s,
				repo: repo,
			}
			return r, nil
		}
	default:
		return nil, fmt.Errorf("unknown step type: %T", step)
	}
}

func (e *StepsRunner) PreTask() *Task {
	return e.createStageTask(StagePre, func(runner StepRunner) *Task {
		return runner.PreTask()
	})
}

func (e *StepsRunner) MainTask() *Task {
	return e.createStageTask(StageMain, func(runner StepRunner) *Task {
		return runner.MainTask()
	})
}

func (e *StepsRunner) PostTask() *Task {
	return e.createStageTask(StagePost, func(runner StepRunner) *Task {
		return runner.PostTask()
	})
}

func (e *StepsRunner) createStageTask(stage Stage, fn func(StepRunner) *Task) *Task {
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

func (e *StepsRunner) executeTasks(stage Stage, tasks []*Task) func(context.Context, *StepRunContext) error {
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
