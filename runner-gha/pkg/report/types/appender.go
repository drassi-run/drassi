/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"context"
	"io"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Appender interface {
	io.Closer
	Append(ctx context.Context, uid string, startAt int, lines []string) error
}

func NewWebsocketAppender(ctx context.Context, wsUrl string, hc *http.Client) (Appender, error) {
	opts := &websocket.DialOptions{
		HTTPClient:      hc,
		CompressionMode: websocket.CompressionContextTakeover,
	}

	conn, _, err := websocket.Dial(ctx, wsUrl, opts)
	if err != nil {
		return nil, err
	}

	a := &wsAppender{
		conn: conn,
	}
	return a, nil
}

type wsAppender struct {
	conn *websocket.Conn
}

// Append log lines into websocket connection.
//
// JobServer: https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/JobServer.cs#L242-L257
// ResultServer: https://github.com/actions/runner/blob/v2.332.0/src/Runner.Common/ResultsServer.cs#L234-L255
func (a *wsAppender) Append(ctx context.Context, uid string, startAt int, lines []string) error {
	data := &LinesWrapper{
		Value:     lines,
		Count:     len(lines),
		StepId:    uid,
		StartLine: startAt,
	}

	return wsjson.Write(ctx, a.conn, data)
}

func (a *wsAppender) Close() error {
	return a.conn.Close(websocket.StatusNormalClosure, "bye")
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
