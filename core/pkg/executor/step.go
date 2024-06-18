package executor

import (
	"context"

	"github.com/dungdm93/drassi/core/pkg/model/workflows"
)

type StepRun interface {
	StepId() string
	Base() *BaseStepRun

	Initialize(ctx context.Context, exec *StepExecutor) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
}

// ensure StepRun implementations
var (
	_ StepRun = (*ScriptStepRun)(nil)
	_ StepRun = (*DockerStepRun)(nil)
	_ StepRun = (*RepositoryStepRun)(nil)
)

type BaseStepRun struct {
	Id               string
	Uid              string
	Name             workflows.Evaluable[string]
	Condition        workflows.Conditional
	ContinueOnError  workflows.Evaluable[bool]
	TimeoutInMinutes workflows.Evaluable[int64]
	Env              workflows.Evaluable[map[string]string]
	Inputs           workflows.Evaluable[map[string]string]
}

func (s *BaseStepRun) StepId() string {
	return s.Id
}

func (s *BaseStepRun) Base() *BaseStepRun {
	return s
}
