package wire_streams

import (
	mock_executor "drassi.run/core/mock/executor"
	mock_command "drassi.run/core/mock/executor/command"
	mock_problem "drassi.run/core/mock/executor/problem"
	mock_reporter "drassi.run/core/mock/executor/reporter"
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

		mw := ProcessCommand(mgr, sup)
		next, err := mw.Handle(line)
		assert.NoError(t, err)
		assert.True(t, next)
	})

	t.Run("process-success", func(t *testing.T) {
		mgr := mock_command.NewMockConsoleManager(ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), line, cmd).Return(nil)

		mw := ProcessCommand(mgr, sup)
		next, err := mw.Handle(line)
		assert.NoError(t, err)
		assert.False(t, next)
	})

	t.Run("process-error", func(t *testing.T) {
		ex := errors.New("process-cmd-error")
		mgr := mock_command.NewMockConsoleManager(ctrl)
		mgr.EXPECT().ParseCommand(line).Return(cmd)
		mgr.EXPECT().Process(t.Context(), line, cmd).Return(ex)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().SetStatus(records.ResultFailure)
		sup.EXPECT().CurrentStep().Return(step)

		mw := ProcessCommand(mgr, sup)
		next, err := mw.Handle(line)
		assert.ErrorIs(t, err, ex)
		assert.False(t, next)
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
	ps   *problemScanner
}

func (s *ScanProblemTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.pm1 = mock_problem.NewMockMatcher(ctrl)
	s.pm2 = mock_problem.NewMockMatcher(ctrl)
	pm := map[string]problem.Matcher{
		"first":  s.pm1,
		"second": s.pm2,
	}
	s.rep = mock_reporter.NewMockReporter(ctrl)
	s.ps = ScanProblem(pm, s.rep).(*problemScanner)
}

func (s *ScanProblemTestSuite) TestNotMatch() {
	t := s.T()
	line := "first line"
	s.pm1.EXPECT().Match(line).Return(nil)
	s.pm2.EXPECT().Match(line).Return(nil)
	next, err := s.ps.Handle(line)
	assert.NoError(t, err)
	assert.True(t, next)
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
	s.rep.EXPECT().AddIssue(iss).Return(nil)

	next, err := s.ps.Handle(line)
	assert.NoError(t, err)
	assert.True(t, next)
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

	next, err := s.ps.Handle(line)
	assert.Error(t, err)
	assert.True(t, next)
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
	s.rep.EXPECT().AddIssue(iss).Return(ex)

	next, err := s.ps.Handle(line)
	assert.ErrorIs(t, err, ex)
	assert.True(t, next)
}
