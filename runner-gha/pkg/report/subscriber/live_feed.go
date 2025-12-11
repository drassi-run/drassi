/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package subscriber

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	xotel "drassi.run/core/util/otel"
	"drassi.run/gha-runner/pkg/log"
	"github.com/coder/websocket"
)

func NewLiveFeedSubscriber(ctx context.Context, wsUrl string, hc *http.Client) (Subscriber, error) {
	opts := &websocket.DialOptions{
		HTTPClient:      hc,
		CompressionMode: websocket.CompressionContextTakeover,
	}

	conn, _, err := websocket.Dial(ctx, wsUrl, opts)
	if err != nil {
		return nil, err
	}

	lf := &liveFeedSubscriber{
		ctx:  ctx,
		conn: conn,
	}
	return lf, nil
}

type liveFeedSubscriber struct {
	ctx  context.Context
	conn *websocket.Conn

	mu sync.Mutex
	wg sync.WaitGroup

	currUid     string
	currBatcher log.Batcher
	lines       atomic.Int64
}

func (s *liveFeedSubscriber) Run(ch <-chan *log.Event) {
	s.wg.Add(1)
	defer s.wg.Done()

	for event := range ch {
		b := s.batcher(event.Uid)
		if u := event.Update; u != nil {
			b.Update(event.Update)
		}
		if event.Kind == log.OnRecordStop {
			_ = b.Close()
		}
	}
}

func (s *liveFeedSubscriber) batcher(uid string) log.Batcher {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currBatcher != nil {
		if s.currUid == uid {
			return s.currBatcher
		}
		_ = s.currBatcher.Close() // batcher of another step
	}

	b := log.NewBatcher()
	s.currUid, s.currBatcher = uid, b

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		s.run(uid, b)
	}()

	return b
}

func (s *liveFeedSubscriber) run(uid string, batcher log.Batcher) {
	ctx, logger := xotel.ChildLogger(s.ctx,
		xotel.ToSlogAttrs(xotel.DrassiStep(uid)),
	)

	for b := range batcher.Channel() {
		data := &linesWrapper{
			Value:     b,
			Count:     len(b),
			StepId:    uid,
			StartLine: s.lines.Load(),
		}
		s.lines.Add(int64(data.Count))

		if payload, err := json.Marshal(data); err != nil {
			logger.Errorf("%v", err)
		} else if err = s.conn.Write(ctx, websocket.MessageText, payload); err != nil {
			logger.Errorf("%v", err)
		} else {
			logger.Infof("successful live feeding log from=[%d-%d)", data.StartLine, data.StartLine+int64(data.Count))
		}
	}
}

func (s *liveFeedSubscriber) Wait() {
	s.wg.Wait()
}

func (s *liveFeedSubscriber) Close() error {
	s.Wait()
	return s.conn.Close(websocket.StatusNormalClosure, "bye")
}

type linesWrapper struct {
	Value     []string `json:"value"`
	Count     int      `json:"count"`
	StepId    string   `json:"step_id"`
	StartLine int64    `json:"start_line"`
}
