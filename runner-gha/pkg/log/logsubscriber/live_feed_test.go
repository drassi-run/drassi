/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package logsubscriber

import (
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	mock_logtypes "drassi.run/gha-runner/mock/log/logtypes"
	"drassi.run/gha-runner/pkg/log"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestLiveFeedSubscriberSuite(t *testing.T) {
	suite.Run(t, new(LiveFeedSubscriberTestSuite))
}

type LiveFeedSubscriberTestSuite struct {
	suite.Suite
	ctrl   *gomock.Controller
	app    *mock_logtypes.MockAppender
	sub    *liveFeedSubscriber
	tmpDir string
}

func (s *LiveFeedSubscriberTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.app = mock_logtypes.NewMockAppender(s.ctrl)
	s.sub = NewLiveFeedSubscriber(s.app).(*liveFeedSubscriber)
	s.tmpDir = s.T().TempDir()
}

func (s *LiveFeedSubscriberTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *LiveFeedSubscriberTestSuite) TestRun() {
	synctest.Test(s.T(), func(t *testing.T) {
		ch := make(chan *log.Event)
		go s.sub.Run(t.Context(), ch)

		uid := "test-uid"
		content := "line 1\nline 2\n"

		logFile := s.tempFile("step.log", content)

		// Send log event
		ch <- &log.Event{
			Uid: uid,
			Update: &log.Update{
				File:   logFile,
				Line:   2,
				Offset: int64(len(content)),
			},
		}

		// Expect Append to be called with lines
		s.app.EXPECT().
			Append(gomock.Any(), uid, 0, []string{"line 1", "line 2"}).
			Return(nil)

		// Close current record to trigger flush
		ch <- &log.Event{
			Uid:  uid,
			Kind: log.OnRecordStop,
		}

		synctest.Wait()

		close(ch)
		s.sub.Wait()

		s.Assert().Equal(2, s.sub.lineCount)
	})
}

func (s *LiveFeedSubscriberTestSuite) TestBatchingByTimeout() {
	synctest.Test(s.T(), func(t *testing.T) {
		ch := make(chan *log.Event)
		go s.sub.Run(t.Context(), ch)

		uid := "test-uid"
		content := "line 1\n"
		logFile := s.tempFile("step.log", content)

		// Expect Append to be called after timeout (1s)
		s.app.EXPECT().
			Append(gomock.Any(), uid, 0, []string{"line 1"}).
			Return(nil)

		ch <- &log.Event{
			Uid: uid,
			Update: &log.Update{
				File:   logFile,
				Line:   1,
				Offset: int64(len(content)),
			},
		}

		// Advance time by more than 1s
		time.Sleep(1100 * time.Millisecond)

		close(ch)
		s.sub.Wait()
		s.Assert().Equal(1, s.sub.lineCount)
	})
}

func (s *LiveFeedSubscriberTestSuite) TestSwitchUid() {
	synctest.Test(s.T(), func(t *testing.T) {
		ch := make(chan *log.Event)
		go s.sub.Run(t.Context(), ch)

		// UID 1
		uid1 := "uid-1"
		content1 := "line from uid-1\n"
		logFile1 := s.tempFile("uid1.log", content1)

		// UID 2
		uid2 := "uid-2"
		content2 := "line from uid-2\n"
		logFile2 := s.tempFile("uid2.log", content2)

		// Switching to uid2 should cause uid1's batcher to close and flush
		s.app.EXPECT().Append(gomock.Any(), uid1, 0, []string{"line from uid-1"}).Return(nil)
		// Eventually uid2 will also flush (on close or timeout)
		s.app.EXPECT().Append(gomock.Any(), uid2, 1, []string{"line from uid-2"}).Return(nil)

		ch <- &log.Event{
			Uid: uid1,
			Update: &log.Update{
				File:   logFile1,
				Line:   1,
				Offset: int64(len(content1)),
			},
		}

		synctest.Wait()

		ch <- &log.Event{
			Uid: uid2,
			Update: &log.Update{
				File:   logFile2,
				Line:   1,
				Offset: int64(len(content2)),
			},
		}

		synctest.Wait()

		// Close record for uid2
		ch <- &log.Event{
			Uid:  uid2,
			Kind: log.OnRecordStop,
		}

		synctest.Wait()

		close(ch)
		s.sub.Wait()
		s.Assert().Equal(2, s.sub.lineCount)
	})
}

func (s *LiveFeedSubscriberTestSuite) TestClose() {
	s.app.EXPECT().Close().Return(nil)
	s.Require().NoError(s.sub.Close())
}

func (s *LiveFeedSubscriberTestSuite) tempFile(name, content string) string {
	f := filepath.Join(s.tmpDir, name)
	err := os.WriteFile(f, []byte(content), 0644)
	s.Require().NoError(err)
	return f
}
