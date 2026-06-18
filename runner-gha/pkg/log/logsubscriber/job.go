/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logsubscriber

import (
	"context"
	"os"
	"sync"

	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/service"
)

////////////// Logs Subscriber for JobService //////////////

func NewJobServiceLogsSubscriber(svc service.JobService) logtypes.Subscriber {
	return &jobServiceLogsSubscriber{
		svc: svc,
		ups: make(map[string]logtypes.Uploader),
	}
}

type jobServiceLogsSubscriber struct {
	svc service.JobService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	ups map[string]logtypes.Uploader
}

func (s *jobServiceLogsSubscriber) Run(ctx context.Context, ch <-chan *log.Event) {
	s.ctx = ctx
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

func (s *jobServiceLogsSubscriber) uploader(uid string) logtypes.Uploader {
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
		xotel.ToSlogAttrs(xotel.Step(e.Uid)),
	)

	u := s.uploader(e.Uid)
	d := e.Update

	f, err := os.Open(d.File)
	if err != nil {
		logger.Errorf("error opening file %s: %v", d.File, err)
		return
	}
	defer f.Close()

	stat := logtypes.NewStat(d.Line, d.Offset)
	if err = u.Upload(ctx, f, stat); err != nil {
		logger.Errorf("error uploading file %s: %s", d.File, err)
	}
}

func (s *jobServiceLogsSubscriber) Wait() {
	s.wg.Wait()
}
