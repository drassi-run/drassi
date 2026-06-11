/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logsubscriber

import (
	"context"
	"sync"
	"time"

	"drassi.run/core/util/context"
	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
)

func NewLiveFeedSubscriber(context xcontext.Provider, app logtypes.Appender) logtypes.Subscriber {
	return &liveFeedSubscriber{
		ctx: context.Context(),
		app: app,
	}
}

type liveFeedSubscriber struct {
	ctx context.Context
	app logtypes.Appender

	mu sync.Mutex
	wg sync.WaitGroup

	currUid     string
	currBatcher log.Batcher
	lineCount   int
}

func (s *liveFeedSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for e := range ch {
		b := s.batcher(e.Uid)
		if u := e.Update; u != nil {
			b.Update(u)
		}
		if e.Kind == log.OnRecordStop {
			s.stopCurrentBatcher()
		}
	}

	// for any reason, OnRecordStop is not received before channel close
	s.stopCurrentBatcher()
}

func (s *liveFeedSubscriber) batcher(uid string) log.Batcher {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currUid == uid {
		return s.currBatcher
	}

	if s.currUid != "" { // batcher of another step
		_ = s.currBatcher.Close()
		s.currUid, s.currBatcher = "", nil
	}

	// In GitHub Runner (C# version)
	// + (batchSize) threshold=100 : https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServerQueue.cs#L372-L389
	// + interval=500ms : https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServerQueue.cs#L33
	b := log.NewBatcher(100, time.Second)
	s.currUid, s.currBatcher = uid, b

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(uid, b)
	}()
	return b
}

func (s *liveFeedSubscriber) stopCurrentBatcher() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currBatcher == nil {
		return
	}

	_ = s.currBatcher.Close()
	s.currUid, s.currBatcher = "", nil
}

func (s *liveFeedSubscriber) run(uid string, batcher log.Batcher) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(uid)),
	)

	for b := range batcher.Channel() {
		if lines, err := b.Scan(); err != nil {
			logger.Errorf("%v", err)
		} else if err = s.app.Append(ctx, uid, s.lineCount, lines); err != nil {
			logger.Errorf("%v", err)
		}
		s.lineCount += b.Lines()
	}
}

func (s *liveFeedSubscriber) Wait() {
	s.wg.Wait()
}

func (s *liveFeedSubscriber) Close() error {
	s.Wait()
	return s.app.Close()
}
