package executor

import "context"

type Supervisor interface {
	Job() JobExecutor
	CurrentStep() StepExecutor
	Context() context.Context

	BeforeJobRun()
	BeforeSandboxCreated()
	BeforeStepRun()
	BeforeTaskRun()

	AfterTaskRun()
	AfterStepRun()
	AfterSandboxTerminated()
	AfterJobRun()
}
