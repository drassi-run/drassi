/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"context"
	"testing"

	mock_executor "drassi.run/core/mock/executor"
	mock_secret "drassi.run/core/mock/executor/secret"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/pkg/model/records"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockStackJob(s *support.Stack, job executor.JobExecutor) {
	run := s.DecorateJobRun(&executor.JobTask{
		Run:      func(ctx context.Context) (*records.Job, error) { return nil, nil },
		Executor: job,
	})
	run(context.Background())
}

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

type consoleHdlCreator func(stack *support.Stack) *command.ConsoleHandler

func testInvalidCommand(creator consoleHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		stack := support.NewStack()
		h := creator(stack)
		cmd := new(command.Command)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	}
}

func testConsoleNoJob(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		stack := support.NewStack()

		h := creator(stack)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testConsoleNoStep(ctrl *gomock.Controller, creator consoleHdlCreator, cmd *command.Command) func(t *testing.T) {
	return func(t *testing.T) {
		stack := support.NewStack()

		h := creator(stack)
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.ErrorIs(t, err, ErrNoStepRunning)
	}
}

func TestConsoleAddPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", testInvalidCommand(ConsoleAddPath))

	cmd := &command.Command{Name: "add-path", Value: "foobar"}
	t.Run("no-job", testConsoleNoJob(ctrl, ConsoleAddPath, cmd))

	t.Run("success", func(t *testing.T) {
		job := mock_executor.NewMockJobExecutor(ctrl)
		stack := support.NewStack()
		mockStackJob(stack, job)

		h := ConsoleAddPath(stack)

		job.EXPECT().AddPath([]string{"foobar"})
		err := command.ConsoleRun(t.Context(), h, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := func(stack *support.Stack) *command.ConsoleHandler {
		return ConsoleSetEnv(stack, nil)
	}

	t.Run("empty-name", testInvalidCommand(creator))

	cmd := &command.Command{Name: "set-env", Params: map[string]string{"name": "XXX"}, Value: "set-env-value"}
	t.Run("no-step", testConsoleNoStep(ctrl, creator, cmd))

	t.Run("success", func(t *testing.T) {
		step := mock_executor.NewMockStepExecutor(ctrl)
		job := mock_executor.NewMockJobExecutor(ctrl)

		stack := support.NewStack()
		mock
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

	t.Run("empty-name", testInvalidCommand(ConsoleSetOutput))

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

	t.Run("empty-name", testInvalidCommand(ConsoleSaveState))

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
