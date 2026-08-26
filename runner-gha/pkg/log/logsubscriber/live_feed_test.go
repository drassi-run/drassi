/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logsubscriber

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	mock_logtypes "drassi.run/gha-runner/mock/log/logtypes"
	"drassi.run/gha-runner/pkg/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestLiveFeedSubscriberSuite(t *testing.T) {
	suite.Run(t, new(LiveFeedSubscriberTestSuite))
}

type LiveFeedSubscriberTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	dir  string
}

func (s *LiveFeedSubscriberTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.dir = s.T().TempDir()
}

func (s *LiveFeedSubscriberTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *LiveFeedSubscriberTestSuite) TestConcurrentSteps() {
	t := s.T()
	app := mock_logtypes.NewMockAppender(s.ctrl)
	sub := NewLiveFeedSubscriber(app)

	ch := make(chan *log.Event, 10)
	ctx := context.Background()

	// Create test log files for step1 and step2
	file1 := filepath.Join(s.dir, "step1.0.log")
	file2 := filepath.Join(s.dir, "step2.0.log")
	require.NoError(t, os.WriteFile(file1, []byte("s1-line1\ns1-line2\n"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("s2-line1\ns2-line2\n"), 0644))

	app.EXPECT().Append(gomock.Any(), "step1", 0, []string{"s1-line1", "s1-line2"}).Return(nil)
	app.EXPECT().Append(gomock.Any(), "step2", 0, []string{"s2-line1", "s2-line2"}).Return(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sub.Run(ctx, ch)
	}()

	// Send OnRecordStart
	ch <- &log.Event{Uid: "step1", Kind: log.OnRecordStart}
	ch <- &log.Event{Uid: "step2", Kind: log.OnRecordStart}

	// Interleaved OnRecordLog
	ch <- &log.Event{
		Uid:  "step1",
		Kind: log.OnRecordLog,
		Update: &log.Update{
			File:     file1,
			Complete: true,
			Offset:   18,
			Line:     2,
		},
	}
	ch <- &log.Event{
		Uid:  "step2",
		Kind: log.OnRecordLog,
		Update: &log.Update{
			File:     file2,
			Complete: true,
			Offset:   18,
			Line:     2,
		},
	}

	// Stop step1
	ch <- &log.Event{Uid: "step1", Kind: log.OnRecordStop}
	// Stop step2
	ch <- &log.Event{Uid: "step2", Kind: log.OnRecordStop}

	close(ch)
	wg.Wait()
}

func (s *LiveFeedSubscriberTestSuite) TestSequentialUpdatesForStep() {
	t := s.T()
	app := mock_logtypes.NewMockAppender(s.ctrl)
	sub := NewLiveFeedSubscriber(app)

	ch := make(chan *log.Event, 10)
	ctx := context.Background()

	file1 := filepath.Join(s.dir, "step1.0.log")
	require.NoError(t, os.WriteFile(file1, []byte("line1\nline2\nline3\n"), 0644))

	app.EXPECT().Append(gomock.Any(), "step1", 0, []string{"line1", "line2", "line3"}).Return(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sub.Run(ctx, ch)
	}()

	ch <- &log.Event{Uid: "step1", Kind: log.OnRecordStart}
	ch <- &log.Event{
		Uid:  "step1",
		Kind: log.OnRecordLog,
		Update: &log.Update{
			File:     file1,
			Complete: true,
			Offset:   18,
			Line:     3,
		},
	}
	ch <- &log.Event{Uid: "step1", Kind: log.OnRecordStop}

	close(ch)
	wg.Wait()
}

func (s *LiveFeedSubscriberTestSuite) TestStopAllSessionsOnChannelClose() {
	t := s.T()
	app := mock_logtypes.NewMockAppender(s.ctrl)
	sub := NewLiveFeedSubscriber(app)

	ch := make(chan *log.Event, 10)
	ctx := context.Background()

	file1 := filepath.Join(s.dir, "step1.0.log")
	require.NoError(t, os.WriteFile(file1, []byte("unstopped-line\n"), 0644))

	app.EXPECT().Append(gomock.Any(), "step1", 0, []string{"unstopped-line"}).Return(nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sub.Run(ctx, ch)
	}()

	ch <- &log.Event{
		Uid:  "step1",
		Kind: log.OnRecordLog,
		Update: &log.Update{
			File:     file1,
			Complete: true,
			Offset:   15,
			Line:     1,
		},
	}

	// Close channel without sending OnRecordStop
	close(ch)
	wg.Wait()
}

func (s *LiveFeedSubscriberTestSuite) TestClose() {
	app := mock_logtypes.NewMockAppender(s.ctrl)
	sub := NewLiveFeedSubscriber(app).(io.Closer)

	app.EXPECT().Close().Return(nil)

	err := sub.Close()
	s.Require().NoError(err)
}
