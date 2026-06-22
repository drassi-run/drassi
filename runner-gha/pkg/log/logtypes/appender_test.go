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
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/stretchr/testify/suite"
)

func TestWebsocketAppenderSuite(t *testing.T) {
	suite.Run(t, new(WebsocketAppenderTestSuite))
}

type WebsocketAppenderTestSuite struct {
	suite.Suite
	server   *httptest.Server
	received chan *LinesWrapper
	accepted atomic.Int32
}

func (s *WebsocketAppenderTestSuite) SetupTest() {
	s.accepted.Store(0)
	s.received = make(chan *LinesWrapper, 10)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.accepted.Add(1)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close(websocket.StatusInternalError, "the sky is falling")

		for {
			var data LinesWrapper
			if err := wsjson.Read(r.Context(), conn, &data); err != nil {
				return
			}
			s.received <- &data
		}
	}))
}

func (s *WebsocketAppenderTestSuite) TearDownTest() {
	s.server.Close()
}

func (s *WebsocketAppenderTestSuite) TestAppend_Success() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 5*time.Second)
	defer cancel()

	wsUrl := strings.Replace(s.server.URL, "http", "ws", 1)

	appender := NewWebsocketAppender(wsUrl, s.server.Client())
	defer appender.Close()
	s.Zero(s.accepted.Load())

	lines := []string{"line1", "line2"}
	err := appender.Append(ctx, "step1", 10, lines)
	s.Require().NoError(err)
	s.Equal(int32(1), s.accepted.Load())

	select {
	case data := <-s.received:
		s.Equal("step1", data.StepId)
		s.Equal(10, data.StartLine)
		s.Equal(lines, data.Value)
		s.Equal(2, data.Count)
	case <-ctx.Done():
		s.Fail("timed out waiting for websocket data")
	}
}

func (s *WebsocketAppenderTestSuite) TestAppend_ChunksLines() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 5*time.Second)
	defer cancel()

	wsUrl := strings.Replace(s.server.URL, "http", "ws", 1)
	appender := NewWebsocketAppender(wsUrl, s.server.Client())
	defer appender.Close()

	lines := make([]string, 2001)
	for i := range lines {
		lines[i] = fmt.Sprintf("line%d", i)
	}

	err := appender.Append(ctx, "step1", 10, lines)
	s.Require().NoError(err)

	expectedStartLines := []int{10, 1010, 2010}
	expectedChunks := [][]string{lines[:1000], lines[1000:2000], lines[2000:]}
	for i := range expectedChunks {
		select {
		case data := <-s.received:
			s.Equal("step1", data.StepId)
			s.Equal(expectedStartLines[i], data.StartLine)
			s.Equal(expectedChunks[i], data.Value)
			s.Equal(len(expectedChunks[i]), data.Count)
		case <-ctx.Done():
			s.Fail("timed out waiting for websocket data")
		}
	}
}

func (s *WebsocketAppenderTestSuite) TestAppend_DialError() {
	appender := NewWebsocketAppender("://invalid", s.server.Client())

	err := appender.Append(s.T().Context(), "step1", 0, []string{"line1"})
	s.Require().ErrorContains(err, "connect to websocket")
}

func (s *WebsocketAppenderTestSuite) TestClose_WithoutConnection() {
	wsUrl := strings.Replace(s.server.URL, "http", "ws", 1)

	appender := NewWebsocketAppender(wsUrl, s.server.Client())
	s.Zero(s.accepted.Load())

	err := appender.Close()
	s.Require().NoError(err)
	s.Zero(s.accepted.Load())
}
