package worker

import (
	"drassi.run/core/pkg/executor"
	"go.uber.org/dig"
)

type ghaListener struct {
	executor.NoopJobListener
	executor.NoopStepListener
}

func (l *ghaListener) OnInitializeJob(exec executor.JobExecutor, scope *dig.Scope) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnRunJob(exec executor.JobExecutor) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnFinalizeJob(exec executor.JobExecutor) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnInitializeStep(exec executor.StepExecutor, scope *dig.Scope) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}

func (l *ghaListener) OnRunStep(exec executor.StepExecutor, stage executor.Stage) executor.EventHandler {
	//TODO implement me
	panic("implement me")
}
