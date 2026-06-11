/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logtypes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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
}

func (s *WebsocketAppenderTestSuite) SetupTest() {
	s.received = make(chan *LinesWrapper, 10)
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	appender, err := NewWebsocketAppender(ctx, wsUrl, s.server.Client())
	s.Require().NoError(err)
	defer appender.Close()

	lines := []string{"line1", "line2"}
	err = appender.Append(ctx, "step1", 10, lines)
	s.Require().NoError(err)

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

func (s *WebsocketAppenderTestSuite) TestClose() {
	ctx := s.T().Context()
	wsUrl := strings.Replace(s.server.URL, "http", "ws", 1)

	appender, err := NewWebsocketAppender(ctx, wsUrl, s.server.Client())
	s.Require().NoError(err)

	err = appender.Close()
	s.Require().NoError(err)
}
