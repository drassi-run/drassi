package subscriber

import (
	"context"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/report"
	"github.com/chainguard-dev/clog"
)

func NewResultServiceStepLogSubscriber(svc *report.ResultService, context xcontext.Provider) Subscriber {
	return &resultServiceStepLogSubscriber{
		svc:  svc,
		ctx:  context.Context(),
		cons: make(map[string]report.Conveyor),
	}
}

type resultServiceStepLogSubscriber struct {
	svc *report.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	cons map[string]report.Conveyor
}

func (s *resultServiceStepLogSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for event := range ch {
		switch event.Kind {
		case log.OnLogRecord:
			c := s.conveyor(event.StepRun)
			c.Update(event.Data)
		case log.OnCompleteStep:
			c := s.conveyor(event.StepRun)
			_ = c.Close()
		}
	}
}

func (s *resultServiceStepLogSubscriber) conveyor(sr executor.StepRun) report.Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	stepId := sr.StepId()
	if c, ok := s.cons[stepId]; ok {
		return c
	}

	c := s.svc.StepLogsConveyor(sr.Base().Uid)
	s.cons[stepId] = c

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(stepId, c)
	}()
	return c
}

func (s *resultServiceStepLogSubscriber) run(stepId string, c report.Conveyor) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(stepId)),
	)

	if r, err := c.Run(ctx); err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("succesful upload step log lines=%d, size=%d", r.Lines, r.Size)
	}
}

func (s *resultServiceStepLogSubscriber) Wait() {
	s.wg.Wait()
}

func NewResultServiceJobLogSubscriber(svc *report.ResultService, context xcontext.Provider) Subscriber {
	return &resultServiceJobLogSubscriber{
		svc: svc,
		ctx: context.Context(),
	}
}

type resultServiceJobLogSubscriber struct {
	svc *report.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	con report.Conveyor
}

func (s *resultServiceJobLogSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for event := range ch {
		if event.Kind == log.OnLogRecord {
			c := s.conveyor()
			c.Update(event.Data)
		}
	}
}

func (s *resultServiceJobLogSubscriber) conveyor() report.Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.con != nil {
		return s.con
	}

	c := s.svc.JobLogsConveyor()
	s.con = c

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(c)
	}()
	return c
}

func (s *resultServiceJobLogSubscriber) run(c report.Conveyor) {
	logger := clog.FromContext(s.ctx)

	if r, err := c.Run(s.ctx); err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("succesful upload job log lines=%d, size=%d", r.Lines, r.Size)
	}
}

func (s *resultServiceJobLogSubscriber) Wait() {
	s.wg.Wait()
}
