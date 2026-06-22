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

	mock_logtypes "drassi.run/gha-runner/mock/log/logtypes"
	mock_service "drassi.run/gha-runner/mock/service"
	"drassi.run/gha-runner/pkg/log"
	"drassi.run/gha-runner/pkg/log/logtypes"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestResultServiceStepLogsSubscriberSuite(t *testing.T) {
	suite.Run(t, new(ResultServiceStepLogsSubscriberTestSuite))
}

type ResultServiceStepLogsSubscriberTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	svc  *mock_service.MockResultService
	sub  *resultServiceStepLogsSubscriber

	tmpDir string

	planUid string
	jobUid  string
}

func (s *ResultServiceStepLogsSubscriberTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())

	s.svc = mock_service.NewMockResultService(s.ctrl)
	s.sub = NewResultServiceStepLogsSubscriber(s.svc).(*resultServiceStepLogsSubscriber)
	s.tmpDir = s.T().TempDir()

	s.planUid = "plan-id"
	s.jobUid = "job-id"
}

func (s *ResultServiceStepLogsSubscriberTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ResultServiceStepLogsSubscriberTestSuite) TestRun() {
	synctest.Test(s.T(), func(t *testing.T) {
		uid := "test-step-uid"
		content := "step log line\n"
		logFile := s.tempFile("step.log", content)

		c := mock_logtypes.NewMockConveyor(s.ctrl)
		s.svc.EXPECT().StepLogsConveyor(uid).Return(c)
		c.EXPECT().Run(gomock.Any()).
			Return(logtypes.NewStat(1, int64(len(content))), nil)

		ch := make(chan *log.Event)
		go s.sub.Run(t.Context(), ch)

		// Record Start
		event := &log.Event{
			Uid:  uid,
			Kind: log.OnRecordStart,
		}
		ch <- event

		// Record Log
		event = &log.Event{
			Uid:  uid,
			Kind: log.OnRecordLog,
			Update: &log.Update{
				File:   logFile,
				Line:   1,
				Offset: int64(len(content)),
			},
		}
		c.EXPECT().Update(event.Update).Times(1)
		ch <- event

		// Record Log complete
		event = &log.Event{
			Uid:  uid,
			Kind: log.OnRecordLog,
			Update: &log.Update{
				File:     logFile,
				Line:     1,
				Offset:   int64(len(content)),
				Complete: true,
			},
		}
		c.EXPECT().Update(event.Update).Times(1)
		ch <- event

		// Record Stop
		c.EXPECT().Close().Return(nil)
		// Now send a stop event
		ch <- &log.Event{
			Uid:  uid,
			Kind: log.OnRecordStop,
		}

		close(ch)
		s.sub.Wait()
	})
}

func (s *ResultServiceStepLogsSubscriberTestSuite) TestConveyorCaching() {
	synctest.Test(s.T(), func(t *testing.T) {
		c := mock_logtypes.NewMockConveyor(s.ctrl)
		c.EXPECT().Run(gomock.Any()).
			Return(new(logtypes.Stat), nil).
			AnyTimes()

		callCount := 0
		s.svc.EXPECT().StepLogsConveyor(gomock.Any()).
			DoAndReturn(func(uid string) logtypes.Conveyor {
				callCount++
				return c
			}).AnyTimes()

		s.sub.ctx = t.Context()
		c1 := s.sub.conveyor("uid1", nil)
		s.Equal(c, c1)
		s.Equal(1, callCount)

		c2 := s.sub.conveyor("uid1", nil)
		s.Equal(c, c2)
		s.Equal(1, callCount, "Should return cached conveyor")

		c3 := s.sub.conveyor("uid2", nil)
		s.Equal(c, c3)
		s.Equal(2, callCount, "Should call provider for new uid")
	})
}

func (s *ResultServiceStepLogsSubscriberTestSuite) tempFile(name, content string) string {
	f := filepath.Join(s.tmpDir, name)
	err := os.WriteFile(f, []byte(content), 0644)
	s.Require().NoError(err)
	return f
}
