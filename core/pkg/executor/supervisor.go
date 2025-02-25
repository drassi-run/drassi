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
	StartContext(ctx context.Context) func()

	Register(...Callback)

	BeforeJobRun(je JobExecutor) error
	AfterJobRun(je JobExecutor, result *records.Job) error

	BeforeStepRun(stage Stage, se StepExecutor) error
	AfterStepRun(stage Stage, se StepExecutor, result *records.Step) error

	BeforeTaskRun(t *Task, se StepExecutor) error
	AfterTaskRun(t *Task, se StepExecutor) error

	ProvideEnv() map[string]string
}

type Callback func(*hook)

type hook struct {
	beforeRunJobCallbacks []func(context.Context, JobExecutor) error
	afterRunJobCallbacks  []func(context.Context, JobExecutor, *records.Job) error

	beforeRunStepCallbacks []func(context.Context, Stage, StepExecutor) error
	afterRunStepCallbacks  []func(context.Context, Stage, StepExecutor, *records.Step) error

	beforeRunTaskCallbacks []func(context.Context, *Task, StepExecutor) error
	afterRunTaskCallbacks  []func(context.Context, *Task, StepExecutor) error

	envProviders []func() map[string]string
}

func BeforeRunJobCallback(cb func(context.Context, JobExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunJobCallbacks = append(h.beforeRunJobCallbacks, cb)
	}
}

func AfterRunJobCallback(cb func(context.Context, JobExecutor, *records.Job) error) Callback {
	return func(h *hook) {
		h.afterRunJobCallbacks = append(h.afterRunJobCallbacks, cb)
	}
}

func BeforeRunStepCallback(cb func(context.Context, Stage, StepExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunStepCallbacks = append(h.beforeRunStepCallbacks, cb)
	}
}

func AfterRunStepCallback(cb func(context.Context, Stage, StepExecutor, *records.Step) error) Callback {
	return func(h *hook) {
		h.afterRunStepCallbacks = append(h.afterRunStepCallbacks, cb)
	}
}

func BeforeRunTaskCallback(cb func(context.Context, *Task, StepExecutor) error) Callback {
	return func(h *hook) {
		h.beforeRunTaskCallbacks = append(h.beforeRunTaskCallbacks, cb)
	}
}

func AfterRunTaskCallback(cb func(context.Context, *Task, StepExecutor) error) Callback {
	return func(h *hook) {
		h.afterRunTaskCallbacks = append(h.afterRunTaskCallbacks, cb)
	}
}

func EnvProvider(fn func() map[string]string) Callback {
	return func(h *hook) {
		h.envProviders = append(h.envProviders, fn)
	}
}

func Env(m map[string]string) Callback {
	fn := func() map[string]string {
		return m
	}

	return func(h *hook) {
		h.envProviders = append(h.envProviders, fn)
	}
}

func NewSupervisor() Supervisor {
	return &supervisor{
		ctx: context.Background(),
	}
}

type supervisor struct {
	hook

	job   JobExecutor
	steps []StepExecutor
	ctx   context.Context
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
	return s.ctx
}

func (s *supervisor) StartContext(ctx context.Context) func() {
	cur := s.ctx
	s.ctx = ctx
	return func() {
		s.ctx = cur
	}
}

func (s *supervisor) Register(callbacks ...Callback) {
	for _, cb := range callbacks {
		cb(&s.hook)
	}
}

func (s *supervisor) BeforeJobRun(je JobExecutor) error {
	s.job = je
	for _, fn := range s.beforeRunJobCallbacks {
		if err := fn(s.ctx, je); err != nil {
			return err
		}
	}

	return nil
}

func (s *supervisor) AfterJobRun(je JobExecutor, result *records.Job) error {
	for _, fn := range s.afterRunJobCallbacks {
		if err := fn(s.ctx, je, result); err != nil {
			return err
		}
	}
	s.job = nil

	return nil
}

func (s *supervisor) BeforeStepRun(stage Stage, se StepExecutor) error {
	s.steps = append(s.steps, se)

	for _, fn := range s.beforeRunStepCallbacks {
		if err := fn(s.ctx, stage, se); err != nil {
			return err
		}
	}
	return nil
}

func (s *supervisor) AfterStepRun(stage Stage, se StepExecutor, result *records.Step) error {
	for _, fn := range s.afterRunStepCallbacks {
		if err := fn(s.ctx, stage, se, result); err != nil {
			return err
		}
	}

	s.steps = s.steps[:len(s.steps)-1]
	return nil
}

func (s *supervisor) BeforeTaskRun(t *Task, se StepExecutor) error {
	for _, fn := range s.beforeRunTaskCallbacks {
		if err := fn(s.ctx, t, se); err != nil {
			return err
		}
	}
	return nil
}

func (s *supervisor) AfterTaskRun(t *Task, se StepExecutor) error {
	for _, fn := range s.afterRunTaskCallbacks {
		if err := fn(s.ctx, t, se); err != nil {
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
