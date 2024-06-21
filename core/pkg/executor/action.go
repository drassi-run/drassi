package executor

import (
	"context"

	"drassi.run/core/pkg/model/actions"
)

type ActionRun interface {
	Initialize(ctx context.Context, exec StepExecutor) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Action() actions.Runs
}

// ensure ActionRun implementations
var (
	_ ActionRun = (*javaScriptActionRun)(nil)
	_ ActionRun = (*dockerActionRun)(nil)
	_ ActionRun = (*compositeActionRun)(nil)
)
