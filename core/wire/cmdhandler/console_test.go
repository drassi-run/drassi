/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_cmdhandler

import (
	mock_executor "drassi.run/core/mock/executor"
	mock_secret "drassi.run/core/mock/executor/secret"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/secret"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestAddSecretMask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		h := AddSecretMask(sm)
		cmd := &command.Command{Name: "add-mask", Value: ""}
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	})

	t.Run("single-secret", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc"))

		h := AddSecretMask(sm)
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

		h := AddSecretMask(sm)
		cmd := &command.Command{Name: "add-mask", Value: "abc\nxyz\r\nfoo  \r  bar"}
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

type consoleHdlCreator func(stack executor.Stack) *command.ConsoleHandler

func testInvalidCommand(ctrl *gomock.Controller, creator consoleHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		stack := mock_executor.NewMockStack(ctrl)
		h := creator(stack)
		cmd := new(command.Command)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	}
}

func testConsoleNoJob(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Job().Return(nil)

		h := creator(stack)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testConsoleNoStep(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Leaf().Return(nil).AnyTimes()
		stack.EXPECT().Root().Return(nil).AnyTimes()
		stack.EXPECT().Stack().Return(nil).AnyTimes()

		h := creator(stack)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoStepRunning)
	}
}

func TestConsoleAddPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", testInvalidCommand(ctrl, ConsoleAddPath))

	cmd := &command.Command{Name: "add-path", Value: "foobar"}
	t.Run("no-job", testConsoleNoJob(ctrl, ConsoleAddPath, cmd))

	t.Run("success", func(t *testing.T) {
		job := mock_executor.NewMockJobExecutor(ctrl)
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Job().Return(job)

		h := ConsoleAddPath(stack)

		job.EXPECT().AddPath([]string{"foobar"})
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := func(stack executor.Stack) *command.ConsoleHandler {
		return ConsoleSetEnv(stack, nil)
	}

	t.Run("empty-name", testInvalidCommand(ctrl, creator))

	cmd := &command.Command{Name: "set-env", Params: map[string]string{"name": "XXX"}, Value: "set-env-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, creator, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)
		job := mock_executor.NewMockJobExecutor(ctrl)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Stack().Return([]executor.StepExecutor{step})
		stack.EXPECT().Job().Return(job)

		h := ConsoleSetEnv(stack, nil)

		step.EXPECT().SetEnv(map[string]string{"XXX": "set-env-value"})
		job.EXPECT().SetEnv(map[string]string{"XXX": "set-env-value"})
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(ctrl, ConsoleSetOutput))

	cmd := &command.Command{Name: "set-output", Params: map[string]string{"name": "XXX"}, Value: "set-output-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, ConsoleSetOutput, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Leaf().Return(step)

		h := ConsoleSetOutput(stack)

		step.EXPECT().SetOutput(map[string]string{"XXX": "set-output-value"})
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(ctrl, ConsoleSaveState))

	cmd := &command.Command{Name: "save-state", Params: map[string]string{"name": "XXX"}, Value: "save-state-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, ConsoleSaveState, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Root().Return(step)

		h := ConsoleSaveState(stack)

		step.EXPECT().SaveState(map[string]string{"XXX": "save-state-value"})
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}
