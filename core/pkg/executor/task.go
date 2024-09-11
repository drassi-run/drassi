package executor

import (
	"drassi.run/core/pkg/model/workflows"
)

type Stage string

const (
	StagePre  Stage = "pre"
	StageMain Stage = "main"
	StagePost Stage = "post"
)

type TaskRun func(StepExecutor) error

type Task struct {
	StepId    string
	Stage     Stage
	Condition workflows.Conditional // default true
	Run       TaskRun
}
