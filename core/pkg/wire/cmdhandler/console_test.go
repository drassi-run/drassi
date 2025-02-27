package wire_cmdhandler

import (
	mock_executor "drassi.run/core/mock/executor"
	mock_logging "drassi.run/core/mock/executor/logging"
	mock_secret "drassi.run/core/mock/executor/secret"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/executor/secret"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestAddSecretMask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	l := mock_logging.NewMockLogger(ctrl)
	l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("empty-value", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		h := AddSecretMask(sm, l)
		cmd := &command.Command{Name: "add-mask", Value: ""}
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	})

	t.Run("single-secret", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc"))

		h := AddSecretMask(sm, l)
		cmd := &command.Command{Name: "add-mask", Value: "abc"}
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})

	t.Run("multi-secret", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc\nxyz\r\nfoo  \r  bar"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("xyz"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("foo"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("bar"))

		h := AddSecretMask(sm, l)
		cmd := &command.Command{Name: "add-mask", Value: "abc\nxyz\r\nfoo  \r  bar"}
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

type consoleHdlCreator func(executor.Supervisor, logging.Logger) *command.ConsoleHandler

func testInvalidCommand(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		sup := mock_executor.NewMockSupervisor(ctrl)
		h := creator(sup, l)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	}
}

func testConsoleNoJob(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Job().Return(nil)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := creator(sup, l)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testConsoleNoStep(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(nil)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := creator(sup, l)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoStepRunning)
	}
}

func TestConsoleAddPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", testInvalidCommand(ctrl, ConsoleAddPath, new(command.Command)))

	cmd := &command.Command{Name: "add-path", Value: "foobar"}
	t.Run("no-job", testConsoleNoJob(ctrl, ConsoleAddPath, cmd))

	t.Run("success", func(t *testing.T) {
		job := mock_executor.NewMockJobExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Job().Return(job)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := ConsoleAddPath(sup, l)

		job.EXPECT().AddPath([]string{"foobar"}).Return(nil)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(ctrl, ConsoleSetEnv, new(command.Command)))

	cmd := &command.Command{Name: "set-env", Params: map[string]string{"name": "XXX"}, Value: "set-env-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, ConsoleSetEnv, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := ConsoleSetEnv(sup, l)

		step.EXPECT().SetEnv(map[string]string{"XXX": "set-env-value"}).Return(nil)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(ctrl, ConsoleSetOutput, new(command.Command)))

	cmd := &command.Command{Name: "set-output", Params: map[string]string{"name": "XXX"}, Value: "set-output-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, ConsoleSetOutput, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := ConsoleSetOutput(sup, l)

		step.EXPECT().SetOutput(map[string]string{"XXX": "set-output-value"}).Return(nil)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(ctrl, ConsoleSaveState, new(command.Command)))

	cmd := &command.Command{Name: "save-state", Params: map[string]string{"name": "XXX"}, Value: "save-state-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, ConsoleSaveState, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)

		l := mock_logging.NewMockLogger(ctrl)
		l.EXPECT().Logf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

		h := ConsoleSaveState(sup, l)

		step.EXPECT().SaveState(map[string]string{"XXX": "save-state-value"}).Return(nil)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}
