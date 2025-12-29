/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package subscriber

import (
	"context"
	"sync"

	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/report"
	"drassi.run/gha-runner/pkg/report/types"
)

////////////// StepLogs Subscriber for ResultService //////////////

func NewResultServiceStepLogsSubscriber(context xcontext.Provider, svc report.ResultService) types.Subscriber {
	return &resultServiceStepLogsSubscriber{
		svc:  svc,
		ctx:  context.Context(),
		cons: make(map[string]types.Conveyor),
	}
}

type resultServiceStepLogsSubscriber struct {
	svc report.ResultService
	ctx context.Context

	mu sync.Mutex
	wg sync.WaitGroup

	cons map[string]types.Conveyor
}

func (s *resultServiceStepLogsSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for event := range ch {
		switch event.Kind {
		case log.OnRecordLog:
			c := s.conveyor(event.Uid)
			c.Update(event.Update)
		case log.OnRecordStop:
			c := s.conveyor(event.Uid)
			if u := event.Update; u != nil {
				c.Update(u)
			}
			_ = c.Close()
		}
	}
}

func (s *resultServiceStepLogsSubscriber) conveyor(uid string) types.Conveyor {
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
		s.run(uid, c)
	}()
	return c
}

func (s *resultServiceStepLogsSubscriber) run(uid string, c types.Conveyor) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(uid)),
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
