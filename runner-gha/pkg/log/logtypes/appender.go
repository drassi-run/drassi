/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/chainguard-dev/clog"
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

	l := clog.FromContext(ctx)
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

		l.Debugf("appending chunk of size=%d to websocket", data.Count)
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

	l := clog.FromContext(ctx)
	opts := &websocket.DialOptions{
		HTTPClient:      a.hc,
		CompressionMode: websocket.CompressionContextTakeover,
	}

	l.Debugf("Dial websocket connection: %s", a.wsUrl)
	conn, resp, err := websocket.Dial(ctx, a.wsUrl, opts)
	if err != nil {
		l.Errorf("websocket.Dial error: %v", err)
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			if c, err := io.ReadAll(resp.Body); err == nil {
				l.Errorf("websocket response body: %s", string(c))
			}
		}
		return err
	}

	a.conn = conn
	return nil
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

func NewFallbackAppender(main Appender, fallback Appender, retryDelay time.Duration) Appender {
	return &fallbackAppender{
		main:       main,
		fallback:   fallback,
		retryDelay: retryDelay,
	}
}

type fallbackAppender struct {
	main     Appender
	fallback Appender

	fbWhen     *time.Time
	retryDelay time.Duration
}

func (a *fallbackAppender) Append(ctx context.Context, uid string, startAt int, lines []string) error {
	usingFb := false
	if when := a.fbWhen; when != nil {
		if time.Since(*when) >= a.retryDelay {
			a.fbWhen = nil // reset
		} else {
			usingFb = true
		}
	}
	if usingFb {
		return a.fallback.Append(ctx, uid, startAt, lines)
	}

	if err := a.main.Append(ctx, uid, startAt, lines); err == nil {
		return nil // main success
	} else {
		clog.ErrorContextf(ctx, "main appender failed, switching to fallback: %v", err)
	}

	// main error => using fallback
	a.fbWhen = new(time.Now())
	return a.fallback.Append(ctx, uid, startAt, lines)
}

func (a *fallbackAppender) Close() error {
	errs := make([]error, 2)
	errs[0] = a.main.Close()
	errs[1] = a.fallback.Close()
	return errors.Join(errs...)
}

type LinesWrapper struct {
	Value     []string `json:"value"`
	Count     int      `json:"count"`
	StepId    string   `json:"step_id"`
	StartLine int      `json:"start_line"`
}
