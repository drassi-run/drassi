package wire_streams

import (
	mock_executor "drassi.run/core/mock/executor"
	mock_command "drassi.run/core/mock/executor/command"
	mock_problem "drassi.run/core/mock/executor/problem"
	mock_reporter "drassi.run/core/mock/executor/reporter"
	mock_secret "drassi.run/core/mock/executor/secret"
	mock_stream "drassi.run/core/mock/stream"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/problem"
	"drassi.run/core/pkg/model/records"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestProcessCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	line := "foobar"
	cmd := &command.Command{Name: "foobar"}

	sup := mock_executor.NewMockSupervisor(ctrl)
	sup.EXPECT().Context().Return(t.Context()).AnyTimes()

	t.Run("non-cmd", func(t *testing.T) {
		mgr := mock_command.NewMockConsoleManager(ctrl)
		mgr.EXPECT().ParseCommand(line).Return(nil)

		hdl := mock_stream.NewMockHandler(ctrl)
		hdl.EXPECT().Handle(t.Context(), line).Return(nil)

		handler := ProcessCommand(mgr, sup)(hdl)
		err := handler.Handle(t.Context(), line)
		assert.NoError(t, err)
	})

	t.Run("process-success", func(t *testing.T) {
		mgr := mock_command.NewMockConsoleManager(ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), line, cmd).Return(nil)

		handler := ProcessCommand(mgr, sup)(nil)
		err := handler.Handle(t.Context(), line)
		assert.NoError(t, err)
	})

	t.Run("process-error", func(t *testing.T) {
		ex := errors.New("process-cmd-error")
		mgr := mock_command.NewMockConsoleManager(ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), line, cmd).Return(ex)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().SetStatus(records.ResultFailure)
		sup.EXPECT().CurrentStep().Return(step)

		handler := ProcessCommand(mgr, sup)(nil)
		err := handler.Handle(t.Context(), line)
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
	rep  *mock_reporter.MockReporter
	hdl  *mock_stream.MockHandler
	ps   *problemScanner
}

func (s *ScanProblemTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.pm1 = mock_problem.NewMockMatcher(s.ctrl)
	s.pm2 = mock_problem.NewMockMatcher(s.ctrl)
	pm := map[string]problem.Matcher{
		"first":  s.pm1,
		"second": s.pm2,
	}
	s.rep = mock_reporter.NewMockReporter(s.ctrl)
	s.hdl = mock_stream.NewMockHandler(s.ctrl)
	s.hdl.EXPECT().Handle(s.T().Context(), gomock.Any()).Return(nil)
	s.ps = ScanProblem(pm, s.rep)(s.hdl).(*problemScanner)
}

func (s *ScanProblemTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *ScanProblemTestSuite) TestNotMatch() {
	t := s.T()
	line := "first line"
	s.pm1.EXPECT().Match(line).Return(nil)
	s.pm2.EXPECT().Match(line).Return(nil)

	err := s.ps.Handle(t.Context(), line)
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
	s.rep.EXPECT().AddIssue(t.Context(), iss).Return(nil)

	err := s.ps.Handle(t.Context(), line)
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

	err := s.ps.Handle(t.Context(), line)
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
	s.rep.EXPECT().AddIssue(t.Context(), iss).Return(ex)

	err := s.ps.Handle(t.Context(), line)
	assert.ErrorIs(t, err, ex)
}

func TestMaskSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	line, maskedLine := "original", "masked"

	sm := mock_secret.NewMockMasker(ctrl)
	sm.EXPECT().Mask(line).Return(maskedLine)

	hdl := mock_stream.NewMockHandler(ctrl)
	hdl.EXPECT().Handle(t.Context(), maskedLine).Return(nil)

	handler := MaskSecret(sm)(hdl)
	err := handler.Handle(t.Context(), line)
	assert.NoError(t, err)
}
