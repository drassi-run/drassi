package executor

import (
	"context"

	"drassi.run/core/pkg/model/workflows"
)

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type Task struct {
	StepId    string
	Stage     Stage
	Condition workflows.Conditional // default true
	Run       func(context.Context, StepExecutor) error
}

type JobRun struct {
	Id   string
	Uid  string
	Name workflows.Evaluable[string]

	Container workflows.Evaluable[*workflows.Container]
	Services  workflows.Evaluable[map[string]*workflows.Container]

	Env      workflows.Evaluable[map[string]string]
	Steps    []StepRun
	Outputs  workflows.Evaluable[map[string]string]
	Defaults workflows.Evaluable[workflows.Defaults]
	// Environment
}
