/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"errors"
	"testing"

	mock_command "drassi.run/core/mock/executor/command"
	mock_problem "drassi.run/core/mock/executor/problem"
	mock_secret "drassi.run/core/mock/executor/secret"
	mock_support "drassi.run/core/mock/executor/support"
	mock_stream "drassi.run/core/mock/stream"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

func TestProcessCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	line := "foobar"
	cmd := &command.Command{Name: "foobar"}

	t.Run("non-cmd", func(t *testing.T) {
		mgr := mock_command.NewMockConsoleManager[any](ctrl)
		mgr.EXPECT().ParseCommand(line).Return(nil)

		hdl := mock_stream.NewMockHandler[any](ctrl)
		hdl.EXPECT().Handle(t.Context(), nil, line).Return(nil)

		handler := ProcessCommand[any](mgr)(hdl)
		err := handler.Handle(t.Context(), nil, line)
		assert.NoError(t, err)
	})

	t.Run("process-success", func(t *testing.T) {
		mgr := mock_command.NewMockConsoleManager[string](ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), "awesome-resource", line, cmd).Return(nil)

		handler := ProcessCommand[string](mgr)(nil)
		err := handler.Handle(t.Context(), "awesome-resource", line)
		assert.NoError(t, err)
	})

	t.Run("process-error", func(t *testing.T) {
		ex := errors.New("process-cmd-error")
		mgr := mock_command.NewMockConsoleManager[any](ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), nil, line, cmd).Return(ex)

		handler := ProcessCommand[any](mgr)(nil)
		err := handler.Handle(t.Context(), nil, line)
		assert.ErrorIs(t, err, ex)
	})
}

func TestScanProblem(t *testing.T) {
	suite.Run(t, new(ScanProblemTestSuite))
}

type ScanProblemTestSuite struct {
	suite.Suite
	ctrl *gomock.Controller
	pm1  *mock_problem.MockMatcher
	pm2  *mock_problem.MockMatcher
	trk  *mock_support.MockTracker
	hdl  *mock_stream.MockHandler[string]
	ps   *problemScanner[string]
	res  string
}

func (s *ScanProblemTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.pm1 = mock_problem.NewMockMatcher(s.ctrl)
	s.pm2 = mock_problem.NewMockMatcher(s.ctrl)
	pm := map[string]problem.Matcher{
		"first":  s.pm1,
		"second": s.pm2,
	}
	s.trk = mock_support.NewMockTracker(s.ctrl)
	s.res = "awesome-resource"
	s.hdl = mock_stream.NewMockHandler[string](s.ctrl)
	s.hdl.EXPECT().Handle(s.T().Context(), s.res, gomock.Any()).Return(nil)
	s.ps = ScanProblem[string](pm, s.trk)(s.hdl).(*problemScanner[string])
}

func (s *ScanProblemTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ScanProblemTestSuite) TestNotMatch() {
	t := s.T()
	line := "first line"
	s.pm1.EXPECT().Match(line).Return(nil)
	s.pm2.EXPECT().Match(line).Return(nil)

	err := s.ps.Handle(t.Context(), s.res, line)
	assert.NoError(t, err)
}

func (s *ScanProblemTestSuite) TestMatchAndSuccess() {
	t := s.T()
	line := "second line"
	pbl := &problem.Problem{
		File:     "/path/to/file",
		Line:     "1",
		Column:   "1",
		Severity: "ERROR",
		Code:     "<code>hello world</code>",
		Message:  "hello from the other side",
	}
	iss, _ := s.ps.toIssuer(pbl)
	s.pm1.EXPECT().Match(line).Return(nil).MinTimes(0).MaxTimes(1)
	s.pm2.EXPECT().Match(line).Return(pbl)
	s.pm1.EXPECT().Reset().Return() // reset other matchers
	s.trk.EXPECT().AddIssue(t.Context(), iss).Return(nil)

	err := s.ps.Handle(t.Context(), s.res, line)
	assert.NoError(t, err)
}

func (s *ScanProblemTestSuite) TestConvertError() {
	t := s.T()
	line := "third line"
	pbl := &problem.Problem{
		Severity: "UNKNOWN",
	}
	s.pm1.EXPECT().Match(line).Return(pbl)
	s.pm2.EXPECT().Match(line).Return(nil).MinTimes(0).MaxTimes(1)
	s.pm2.EXPECT().Reset().Return()

	err := s.ps.Handle(t.Context(), s.res, line)
	assert.Error(t, err)
}

func (s *ScanProblemTestSuite) TestReportError() {
	t := s.T()
	line := "forth line"
	pbl := &problem.Problem{
		File:     "/path/to/file",
		Line:     "1",
		Column:   "1",
		Severity: "ERROR",
		Code:     "<code>hello world</code>",
		Message:  "hello from the other side",
	}
	iss, _ := s.ps.toIssuer(pbl)
	ex := errors.New("report-error")
	s.pm1.EXPECT().Match(line).Return(pbl)
	s.pm2.EXPECT().Match(line).Return(nil).MinTimes(0).MaxTimes(1)
	s.pm2.EXPECT().Reset().Return()
	s.trk.EXPECT().AddIssue(t.Context(), iss).Return(ex)

	err := s.ps.Handle(t.Context(), s.res, line)
	assert.ErrorIs(t, err, ex)
}

func TestMaskSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	line, maskedLine := "original", "masked"

	sm := mock_secret.NewMockMasker(ctrl)
	sm.EXPECT().Mask(line).Return(maskedLine)

	hdl := mock_stream.NewMockHandler[string](ctrl)
	hdl.EXPECT().Handle(t.Context(), "res", maskedLine).Return(nil)

	handler := MaskSecret[string](sm)(hdl)
	err := handler.Handle(t.Context(), "res", line)
	assert.NoError(t, err)
}
