/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package subscriber

import (
	"context"
	"os"
	"sync"

	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/report"
	"drassi.run/gha-runner/pkg/report/types"
)

////////////// Logs Subscriber for JobService //////////////

func NewJobServiceLogsSubscriber(context xcontext.Provider, svc *report.JobService) types.Subscriber {
	return &jobServiceLogsSubscriber{
		svc: svc,
		ctx: context.Context(),
		ups: make(map[string]types.Uploader),
	}
}

type jobServiceLogsSubscriber struct {
	svc *report.JobService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	ups map[string]types.Uploader
}

func (s *jobServiceLogsSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for e := range ch {
		if e.Update == nil || !e.Update.Complete || e.Update.Offset == 0 {
			continue
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()

			s.handle(e)
		}()
	}
}

func (s *jobServiceLogsSubscriber) uploader(uid string) types.Uploader {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.ups[uid]; ok {
		return c
	}

	u := s.svc.LogsUploader(uid)
	s.ups[uid] = u
	return u
}

func (s *jobServiceLogsSubscriber) handle(e *log.Event) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(e.Uid)),
	)

	u := s.uploader(e.Uid)
	d := e.Update

	f, err := os.Open(d.File)
	if err != nil {
		logger.Errorf("error opening file %s: %v", d.File, err)
		return
	}
	defer f.Close()

	stat := types.NewStat(d.Line, d.Offset)
	if err = u.Upload(ctx, f, stat); err != nil {
		logger.Errorf("error uploading file %s: %s", d.File, err)
	}
}

func (s *jobServiceLogsSubscriber) Wait() {
	s.wg.Wait()
}
