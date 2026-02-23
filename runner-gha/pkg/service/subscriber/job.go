package subscriber

import (
	"context"
	"os"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/service"
)

func NewJobServiceLogSubscriber(svc *service.JobService, context xcontext.Provider) Subscriber {
	return &jobServiceLogSubscriber{
		svc: svc,
		ctx: context.Context(),
		ups: make(map[string]service.Uploader),
	}
}

type jobServiceLogSubscriber struct {
	svc *service.JobService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	ups map[string]service.Uploader
}

func (s *jobServiceLogSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for e := range ch {
		if e.Data == nil || e.Data.Status != log.FileClose || e.Data.Size == 0 {
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			s.handle(e)
		}()
	}
}

func (s *jobServiceLogSubscriber) uploader(sr executor.StepRun) service.Uploader {
	s.mu.Lock()
	defer s.mu.Unlock()

	recordId := "" // TODO
	if c, ok := s.ups[recordId]; ok {
		return c
	}

	u := s.svc.LogUploader(recordId)
	s.ups[recordId] = u
	return u
}

func (s *jobServiceLogSubscriber) handle(e *log.Event) {
	stepId := e.StepRun.StepId()
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(stepId)),
	)

	u := s.uploader(e.StepRun)
	d := e.Data

	f, err := os.Open(d.File)
	if err != nil {
		logger.Errorf("error opening file %s: %v", d.File, err)
		return
	} else {
		defer f.Close()
	}

	stat := service.NewStat(d.Line, d.Size)
	if err = u.Upload(ctx, f, stat); err != nil {
		logger.Errorf("error uploading file %s: %s", d.File, err)
	}
}

func (s *jobServiceLogSubscriber) Wait() {
	s.wg.Wait()
}
