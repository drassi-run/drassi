package executor

import (
	"drassi.run/core/pkg/model/workflows"

	"go.uber.org/dig"
)

type NodeStepRun struct {
	BaseStepRun

	Runtime string
	Main    string

	Pre   string
	PreIf workflows.Conditional

	Post   string
	PostIf workflows.Conditional
}

func (sr *NodeStepRun) Initialize(exec StepExecutor, scope *dig.Scope) error {
	// TODO copy to sandbox
	return nil
}

func (sr *NodeStepRun) PreTask() *Task {
	if sr.Pre == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := sr.PreIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePre,
		Condition: condition,
		Run:       sr.execute(StagePre),
	}
}

func (sr *NodeStepRun) MainTask() *Task {
	return &Task{
		StepId:    sr.Id,
		Stage:     StageMain,
		Condition: sr.Condition,
		Run:       sr.execute(StageMain),
	}
}

func (sr *NodeStepRun) PostTask() *Task {
	if sr.Post == "" {
		return nil
	}
	// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/ActionManifestManager.cs#L451-L471
	condition := sr.PostIf
	if condition == "" {
		condition = "always()"
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePost,
		Condition: condition,
		Run:       sr.execute(StagePost),
	}
}

func (sr *NodeStepRun) execute(stage Stage) TaskRun {
	return nil
}
