/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logsubscriber

import (
	"context"
	"sync"

	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"drassi.run/gha-runner/pkg/service"
	"go.opentelemetry.io/otel/attribute"
)

////////////// StepLogs Subscriber for ResultService //////////////

func NewResultServiceStepLogsSubscriber(svc service.ResultService) logtypes.Subscriber {
	return &resultServiceStepLogsSubscriber{
		svc:  svc,
		cons: make(map[string]logtypes.Conveyor),
	}
}

type resultServiceStepLogsSubscriber struct {
	svc service.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	cons map[string]logtypes.Conveyor
}

func (s *resultServiceStepLogsSubscriber) Run(ctx context.Context, ch <-chan *log.Event) {
	s.ctx = ctx
	s.wg.Add(1)
	defer s.wg.Done()

	for event := range ch {
		switch event.Kind {
		case log.OnRecordLog:
			c := s.conveyor(event.Uid, event.Attrs)
			c.Update(event.Update)
		case log.OnRecordStop:
			c := s.conveyor(event.Uid, event.Attrs)
			if u := event.Update; u != nil {
				c.Update(u)
			}
			_ = c.Close()
		}
	}
}

func (s *resultServiceStepLogsSubscriber) conveyor(uid string, attrs []attribute.KeyValue) logtypes.Conveyor {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.cons[uid]; ok {
		return c
	}

	c := s.svc.StepLogsConveyor(uid)
	s.cons[uid] = c

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(c, attrs)
	}()
	return c
}

func (s *resultServiceStepLogsSubscriber) run(c logtypes.Conveyor, attrs []attribute.KeyValue) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(attrs...),
	)

	if r, err := c.Run(ctx); err != nil {
		logger.Errorf("%v", err)
	} else {
		logger.Infof("successful upload step log lines=%d, size=%d", r.Lines, r.Size)
	}
}

func (s *resultServiceStepLogsSubscriber) Wait() {
	s.wg.Wait()
}
