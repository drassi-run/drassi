package executor

import (
	"context"

	"github.com/dungdm93/drasi/pkg/model/actions"
)

type ActionRunner interface {
	Initialize(ctx context.Context) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
	Action() actions.Runs
}
