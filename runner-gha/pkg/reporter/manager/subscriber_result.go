package manager

import (
	"context"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/reporter/service"
	"github.com/chainguard-dev/clog"
)

type Subscriber interface {
	Run(ch <-chan *Event)
	Wait()
}

func NewResultStepLogListener(svc *service.ResultService, context xcontext.Provider) Subscriber {
	return &resultStepLogSubscriber{
		svc:  svc,
		ctx:  context.Context(),
		cons: make(map[string]service.Conveyor),
	}
}

type resultStepLogSubscriber struct {
	svc *service.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	cons map[string]service.Conveyor
}

func (s *resultStepLogSubscriber) Run(ch <-chan *Event) {
	for event := range ch {
		switch event.Kind {
		case OnLogRecord:
			c := s.conveyor(event.StepRun)
			c.Update(event.Data)
		case OnCompleteStep:
			c := s.conveyor(event.StepRun)
			_ = c.Close()
		}
	}
}

func (s *resultStepLogSubscriber) conveyor(sr executor.StepRun) service.Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	stepId := sr.StepId()
	if c, ok := s.cons[stepId]; ok {
		return c
	}

	c := s.svc.StepLogsConveyor(sr)
	s.cons[stepId] = c

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(stepId, c)
	}()
	return c
}

func (s *resultStepLogSubscriber) run(stepId string, c service.Conveyor) {
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

func NewResultJobLogListener(svc *service.ResultService, context xcontext.Provider) Subscriber {
	return &resultJobLogSubscriber{
		svc: svc,
		ctx: context.Context(),
	}
}

type resultJobLogSubscriber struct {
	svc *service.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	con service.Conveyor
}

func (s *resultJobLogSubscriber) Run(ch <-chan *Event) {
	for event := range ch {
		switch event.Kind {
		case OnLogRecord:
			c := s.conveyor()
			c.Update(event.Data)
		}
	}
}

func (s *resultJobLogSubscriber) conveyor() service.Conveyor {
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

func (s *resultJobLogSubscriber) run(c service.Conveyor) {
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
