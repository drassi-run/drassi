package service

import (
	"context"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/reporter"
	"drassi.run/gha-runner/pkg/reporter/log"
	"github.com/chainguard-dev/clog"
)

func NewResultStepLogListener(svc *ResultService, context xcontext.Provider) reporter.Subscriber {
	return &resultStepLogSubscriber{
		svc:  svc,
		ctx:  context.Context(),
		cons: make(map[string]Conveyor),
	}
}

type resultStepLogSubscriber struct {
	svc *ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	cons map[string]Conveyor
}

func (s *resultStepLogSubscriber) OnNewStep(sr executor.StepRun) {
}

func (s *resultStepLogSubscriber) OnLogRecord(sr executor.StepRun, update *log.Update) {
	c := s.conveyor(sr)
	c.Update(update)
}

func (s *resultStepLogSubscriber) OnCompleteStep(sr executor.StepRun) {
	c := s.conveyor(sr)
	c.Close()
}

func (s *resultStepLogSubscriber) conveyor(sr executor.StepRun) Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	stepId := sr.StepId()
	if c, ok := s.cons[stepId]; ok {
		return c
	}

	c := s.svc.StepLogsConveyor(sr)
	s.cons[stepId] = c

	s.wg.Go(func() {
		s.run(stepId, c)
	})
	return c
}

func (s *resultStepLogSubscriber) run(stepId string, c Conveyor) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(stepId)),
	)

	if r, err := c.Run(ctx); err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("succesful upload step log lines=%d, size=%d", r.Lines, r.Size)
	}
}

func (s *resultStepLogSubscriber) Wait() {
	s.wg.Wait()
}

func NewResultJobLogListener(svc *ResultService, context xcontext.Provider) reporter.Subscriber {
	return &resultJobLogSubscriber{
		svc: svc,
		ctx: context.Context(),
	}
}

type resultJobLogSubscriber struct {
	svc *ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	con Conveyor
}

func (s *resultJobLogSubscriber) OnNewStep(_ executor.StepRun) {}

func (s *resultJobLogSubscriber) OnLogRecord(_ executor.StepRun, update *log.Update) {
	c := s.conveyor()
	c.Update(update)
}

func (s *resultJobLogSubscriber) OnCompleteStep(_ executor.StepRun) {}

func (s *resultJobLogSubscriber) conveyor() Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.con != nil {
		return s.con
	}

	c := s.svc.JobLogsConveyor()
	s.con = c

	s.wg.Go(func() {
		s.run(c)
	})
	return c
}

func (s *resultJobLogSubscriber) run(c Conveyor) {
	logger := clog.FromContext(s.ctx)

	if r, err := c.Run(s.ctx); err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("succesful upload job log lines=%d, size=%d", r.Lines, r.Size)
	}
}

func (s *resultJobLogSubscriber) Wait() {
	s.wg.Wait()
}
