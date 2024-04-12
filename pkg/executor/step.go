package executor

import "context"

type StepRunner interface {
	Initialize(ctx context.Context) error
	PreTask() *Task
	MainTask() *Task
	PostTask() *Task
}
