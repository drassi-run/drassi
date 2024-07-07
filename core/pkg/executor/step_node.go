package executor

import (
	"context"

	"drassi.run/core/pkg/model/dossiers"
	"drassi.run/core/pkg/model/workflows"
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

func (sr *NodeStepRun) SetContextInfo(dossier *dossiers.Dossier) {
}

func (sr *NodeStepRun) Initialize(ctx context.Context, exec StepExecutor) error {
	// TODO copy to sandbox
	return nil
}

func (sr *NodeStepRun) PreTask() *Task {
	if sr.Pre == "" {
		return nil
	}
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePre,
		Condition: sr.PreIf,
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
	return &Task{
		StepId:    sr.Id,
		Stage:     StagePost,
		Condition: sr.PostIf,
		Run:       sr.execute(StagePost),
	}
}

func (sr *NodeStepRun) execute(stage Stage) func(context.Context, StepExecutor) error {
	return nil
}
