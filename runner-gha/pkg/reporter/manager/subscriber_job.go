package manager

import (
	"context"
	"os"
	"sync"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/reporter/log"
	"drassi.run/gha-runner/pkg/reporter/service"
)

type jobLogSubscriber struct {
	svc *service.JobService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	ups map[string]service.Uploader
}

func (s *jobLogSubscriber) Run(ch <-chan *Event) {
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

func (s *jobLogSubscriber) uploader(sr executor.StepRun) service.Uploader {
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

func (s *jobLogSubscriber) handle(e *Event) {
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

func (s *jobLogSubscriber) Wait() {
	s.wg.Wait()
}
