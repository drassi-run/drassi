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
	steps   []workflows.Step
	runners []StepRunner
	stepMap map[string]workflows.Step
}

func (e *StepsRunner) Initialize(ctx context.Context) (err error) {
	e.stepMap = make(map[string]workflows.Step, len(e.steps))
	e.runners = make([]StepRunner, len(e.steps))
	for i, step := range e.steps {
		e.stepMap[step.Base().Id] = step
		e.runners[i], err = e.createStepRunner(step)
		if err != nil {
			return
		}
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, runner := range e.runners {
		r := runner
		g.Go(func() error {
			return r.Initialize(ctx)
		})
	}
	return g.Wait()
}

func (e *StepsRunner) createStepRunner(step workflows.Step) (StepRunner, error) {
	switch s := step.(type) {
	case *workflows.RunStep:
		r := &runStepRunner{
			rCtx: e.rCtx, // TODO .Clone()
			step: s,
		}
		return r, nil
	case *workflows.UsesStep:
		if image, ok := strings.CutPrefix(s.Uses, "docker://"); ok {
			r := &usesDockerStepRunner{
				rCtx:  e.rCtx, // TODO .Clone()
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
				rCtx: e.rCtx, // TODO .Clone()
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
			_ = rCtx.runStep(ctx, task)
		}
		return nil
	}
}
