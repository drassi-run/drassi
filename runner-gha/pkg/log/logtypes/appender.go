/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Appender interface {
	Append(ctx context.Context, uid string, startAt int, lines []string) error
	Close() error
}

func NewWebsocketAppender(wsUrl string, hc *http.Client) Appender {
	return &wsAppender{
		wsUrl: wsUrl,
		hc:    hc,
	}
}

// wsAppender implement Appender that write log lines into websocket connection.
//
// JobServer: https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServer.cs#L242-L257
// ResultServer: https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/ResultsServer.cs#L234-L255
type wsAppender struct {
	wsUrl string
	hc    *http.Client

	mu   sync.Mutex
	conn *websocket.Conn
}

func (a *wsAppender) Append(ctx context.Context, uid string, startAt int, lines []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if err := a.connect(ctx); err != nil {
		return fmt.Errorf("connect to websocket: %w", err)
	}

	// actions/runner (C#) process at most about 500 lines once
	// https://github.com/actions/runner/blob/v2.335.1/src/Runner.Common/JobServerQueue.cs#L362
	for chunk := range slices.Chunk(lines, 1000) {
		data := &LinesWrapper{
			Value:     chunk,
			Count:     len(chunk),
			StepId:    uid,
			StartLine: startAt,
		}
		startAt += data.Count

		if err := wsjson.Write(ctx, a.conn, data); err != nil {
			return fmt.Errorf("append logs to websocket: %w", err)
		}
	}
	return nil
}

func (a *wsAppender) connect(ctx context.Context) error {
	if a.conn != nil {
		return nil
	}

	opts := &websocket.DialOptions{
		HTTPClient:      a.hc,
		CompressionMode: websocket.CompressionContextTakeover,
	}
	if conn, resp, err := websocket.Dial(ctx, a.wsUrl, opts); err != nil {
		return err
	} else {
		if body := resp.Body; body != nil {
			defer body.Close()
		}
		a.conn = conn
		return nil
	}
}

func (a *wsAppender) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn == nil {
		return nil
	}

	err := a.conn.Close(websocket.StatusNormalClosure, "bye")
	a.conn = nil
	return err
}

type FuncAppender func(ctx context.Context, uid string, startAt int, lines []string) error

func (f FuncAppender) Append(ctx context.Context, uid string, startAt int, lines []string) error {
	return f(ctx, uid, startAt, lines)
}

func (f FuncAppender) Close() error { return nil }

type LinesWrapper struct {
	Value     []string `json:"value"`
	Count     int      `json:"count"`
	StepId    string   `json:"step_id"`
	StartLine int      `json:"start_line"`
}
