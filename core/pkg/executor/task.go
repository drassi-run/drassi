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

type TaskRun func(context.Context, StepExecutor) error

type Task struct {
	StepId    string
	Stage     Stage
	Condition workflows.Conditional
	Run       TaskRun
}
