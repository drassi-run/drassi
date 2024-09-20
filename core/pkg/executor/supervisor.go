package executor

import (
	"context"
	"maps"

	"drassi.run/core/pkg/model/records"
)

type Supervisor interface {
	Job() JobExecutor
	CurrentStep() StepExecutor
	Context() context.Context

	Register(...Callback)

	BeforeJobRun(je JobExecutor) error
	AfterJobRun(je JobExecutor, result *records.Job) error

	BeforeStepRun(se StepExecutor) error
	AfterStepRun(se StepExecutor, result *records.Step) error

	BeforeTaskRun(t *Task, se StepExecutor) error
	AfterTaskRun(t *Task, se StepExecutor) error

	ProvideEnv() map[string]string
}

type Callback func(*hook)

type hook struct {
	beforeRunJobCallbacks []func(JobExecutor) error
	afterRunJobCallbacks  []func(JobExecutor, *records.Job) error

	beforeRunStepCallbacks []func(StepExecutor) error
	afterRunStepCallbacks  []func(StepExecutor, *records.Step) error

	beforeRunTaskCallbacks []func(*Task, StepExecutor) error
	afterRunTaskCallbacks  []func(*Task, StepExecutor) error

	envProviders []func() map[string]string
}

func BeforeRunJobCallback(cb func(JobExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunJobCallbacks = append(h.beforeRunJobCallbacks, cb)
	}
}

func AfterRunJobCallback(cb func(JobExecutor, *records.Job) error) Callback {
	return func(h *hook) {
		h.afterRunJobCallbacks = append(h.afterRunJobCallbacks, cb)
	}
}

func BeforeRunStepCallback(cb func(StepExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunStepCallbacks = append(h.beforeRunStepCallbacks, cb)
	}
}

func AfterRunStepCallback(cb func(StepExecutor, *records.Step) error) Callback {
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

func NewSupervisor() Supervisor {
	return new(supervisor)
}

type supervisor struct {
	hook

	job   JobExecutor
	steps []StepExecutor
}

func (s *supervisor) Job() JobExecutor {
	return s.job
}

func (s *supervisor) CurrentStep() StepExecutor {
	if len(s.steps) == 0 {
		return nil
	}
	return s.steps[len(s.steps)-1]
}

func (s *supervisor) Context() context.Context {
	if len(s.steps) > 0 {
		step := s.CurrentStep()
		return step.Context()
	}
	if s.job != nil {
		return s.job.Context()
	}
	return context.Background()
}

func (s *supervisor) Register(callbacks ...Callback) {
	for _, cb := range callbacks {
		cb(&s.hook)
	}
}

func (s *supervisor) BeforeJobRun(je JobExecutor) error {
	s.job = je
	for _, cb := range s.beforeRunJobCallbacks {
		if err := cb(je); err != nil {
			return err
		}
	}

	return nil
}

func (s *supervisor) AfterJobRun(je JobExecutor, result *records.Job) error {
	for _, cb := range s.afterRunJobCallbacks {
		if err := cb(je, result); err != nil {
			return err
		}
	}
	s.job = nil

	return nil
}

func (s *supervisor) BeforeStepRun(se StepExecutor) error {
	s.steps = append(s.steps, se)

	for _, cb := range s.beforeRunStepCallbacks {
		if err := cb(se); err != nil {
			return err
		}
	}
	return nil
}

func (s *supervisor) AfterStepRun(se StepExecutor, result *records.Step) error {
	for _, cb := range s.afterRunStepCallbacks {
		if err := cb(se, result); err != nil {
			return err
		}
	}

	s.steps = s.steps[:len(s.steps)-1]
	return nil
}

func (s *supervisor) BeforeTaskRun(t *Task, se StepExecutor) error {
	for _, cb := range s.beforeRunTaskCallbacks {
		if err := cb(t, se); err != nil {
			return err
		}
	}
	return nil
}

func (s *supervisor) AfterTaskRun(t *Task, se StepExecutor) error {
	for _, cb := range s.afterRunTaskCallbacks {
		if err := cb(t, se); err != nil {
			return err
		}
	}
	return nil
}

func (s *supervisor) ProvideEnv() map[string]string {
	m := make(map[string]string)
	for _, ep := range s.hook.envProviders {
		env := ep()
		maps.Copy(m, env)
	}
	return m
}
