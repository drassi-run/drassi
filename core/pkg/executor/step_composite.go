package executor

import (
	"fmt"
	"slices"

	"drassi.run/core/pkg/model/dossiers"
	"go.uber.org/dig"
)

type CompositeStepRun struct {
	BaseStepRun

	StepRuns []StepRun
}

func (sr *CompositeStepRun) Initialize(exec StepExecutor, scope *dig.Scope) error {
	// TODO: concurrent version of Initialize is temporary disable because of concurrent map writes in scope
	//g, ctx := errgroup.WithContext(exec.Context())
	//for _, step := range sr.StepRuns {
	//	cExec := exec.NewChildExecutor(step)
	//	cScope := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
	//	g.Go(func() error {
	//		cExec.SetContext(ctx)
	//		return cExec.Initialize(cScope)
	//	})
	//}
	//return g.Wait()

	ctx := exec.Context()
	for _, step := range sr.StepRuns {
		cExec := exec.NewChildExecutor(step)
		cScope := scope.Scope(fmt.Sprintf("step(%s)", exec.StepId()))
		cExec.SetContext(ctx)
		if err := cExec.Initialize(cScope); err != nil {
			return err
		}
	}
	return nil
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
		Run: func(exec StepExecutor) error {
			for _, id := range taskIds {
				cExec := exec.ChildExecutor(id)
				if cExec == nil {
					return fmt.Errorf(`task %q has no child context`, id)
				}

				cExec.SetContext(exec.Context())
				res := cExec.RunStep(fn)
				if res != nil && res.Conclusion == dossiers.ResultFailure {
					return fmt.Errorf(`step %q (%s) failed`, id, stage)
				}
			}
			return nil
		},
	}
}
