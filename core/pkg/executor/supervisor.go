package executor

import "context"

type Supervisor interface {
	Job() JobExecutor
	CurrentStep() StepExecutor
	Context() context.Context

	Register(Callback)

	BeforeJobRun(je JobExecutor)
	AfterJobRun(je JobExecutor, output map[string]string)

	BeforeStepRun(se StepExecutor)
	AfterStepRun(se StepExecutor)

	BeforeTaskRun(task *Task, se StepExecutor) error
	AfterTaskRun(task *Task, se StepExecutor) error

	ProvideEnv() map[string]string
}

type Callback func(*hook)

type hook struct {
	beforeRunJobCallbacks []func(JobExecutor) error
	afterRunJobCallbacks  []func(JobExecutor, map[string]string) error

	beforeRunStepCallbacks []func(StepExecutor) error
	afterRunStepCallbacks  []func(StepExecutor) error

	beforeRunTaskCallbacks []func(*Task, StepExecutor) error
	afterRunTaskCallbacks  []func(*Task, StepExecutor) error

	envProviders []func() map[string]string
}

func BeforeRunJobCallback(cb func(JobExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunJobCallbacks = append(h.beforeRunJobCallbacks, cb)
	}
}

func AfterRunJobCallback(cb func(JobExecutor, map[string]string) error) Callback {
	return func(h *hook) {
		h.afterRunJobCallbacks = append(h.afterRunJobCallbacks, cb)
	}
}

func BeforeRunStepCallback(cb func(executor StepExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunStepCallbacks = append(h.beforeRunStepCallbacks, cb)
	}
}

func AfterRunStepCallback(cb func(executor StepExecutor) error) Callback {
	return func(h *hook) {
		h.afterRunStepCallbacks = append(h.afterRunStepCallbacks, cb)
	}
}

func BeforeRunTaskCallback(cb func(*Task, StepExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunTaskCallbacks = append(h.beforeRunTaskCallbacks, cb)
	}
}

func AfterRunTaskCallback(cb func(*Task, StepExecutor) error) Callback {
	return func(h *hook) {
		h.afterRunTaskCallbacks = append(h.afterRunTaskCallbacks, cb)
	}
}

func EnvProvider(fn func() map[string]string) Callback {
	return func(h *hook) {
		h.envProviders = append(h.envProviders, fn)
	}
}
