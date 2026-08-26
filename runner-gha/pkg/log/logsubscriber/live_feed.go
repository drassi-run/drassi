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

	"drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"go.opentelemetry.io/otel/attribute"
)

func NewLiveFeedSubscriber(app logtypes.Appender) logtypes.Subscriber {
	return &liveFeedSubscriber{
		app:      app,
		sessions: make(map[string]*liveFeedSession),
	}
}

type liveFeedSession struct {
	log.Batcher
	uid       string
	lineCount int
}

type liveFeedSubscriber struct {
	ctx context.Context
	app logtypes.Appender

	mu sync.Mutex
	wg sync.WaitGroup

	sessions map[string]*liveFeedSession
}

func (s *liveFeedSubscriber) Run(ctx context.Context, ch <-chan *log.Event) {
	s.ctx = ctx

	for e := range ch {
		sess := s.session(e.Uid, e.Attrs)
		if u := e.Update; u != nil {
			sess.Update(u)
		}
		if e.Kind == log.OnRecordStop {
			s.stopSession(e.Uid)
		}
	}

	// for any reason, OnRecordStop is not received before channel close
	s.stopAllSessions()
	s.wg.Wait()
}

func (s *liveFeedSubscriber) session(uid string, attrs []attribute.KeyValue) *liveFeedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.sessions[uid]; ok {
		return sess
	}

	b := log.NewBatcher(100, time.Second)
	sess := &liveFeedSession{
		Batcher: b,
		uid:     uid,
	}
	s.sessions[uid] = sess

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.run(sess, attrs)
	}()
	return sess
}

func (s *liveFeedSubscriber) stopSession(uid string) {
	s.mu.Lock()
	sess, ok := s.sessions[uid]
	if ok {
		delete(s.sessions, uid)
	}
	s.mu.Unlock()

	if ok {
		_ = sess.Close()
	}
}

func (s *liveFeedSubscriber) stopAllSessions() {
	s.mu.Lock()
	all := make([]*liveFeedSession, 0, len(s.sessions))
	for _, sess := range s.sessions {
		all = append(all, sess)
	}
	clear(s.sessions)
	s.mu.Unlock()

	for _, sess := range all {
		_ = sess.Close()
	}
}

func (s *liveFeedSubscriber) run(sess *liveFeedSession, attrs []attribute.KeyValue) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(attrs...),
	)

	for b := range sess.Channel() {
		if lines, err := b.Scan(); err != nil {
			logger.Errorf("scan batch error: %v", err)
		} else {
			logger.Debugf("LiveFeed - streaming %d log lines", len(lines))
			if err = s.app.Append(ctx, sess.uid, sess.lineCount, lines); err != nil {
				logger.Errorf("LiveFeed - stream logs error: %v", err)
			} else {
				logger.Debugf("LiveFeed - streamed %d lines", len(lines))
			}
		}
		sess.lineCount += b.Lines()
	}
}

func (s *liveFeedSubscriber) Close() error {
	return s.app.Close()
}
