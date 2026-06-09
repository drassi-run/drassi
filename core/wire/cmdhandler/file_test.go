/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_cmdhandler

import (
	"archive/tar"
	"context"
	"io"
	"strings"
	"testing"

	mock_executor "drassi.run/core/mock/executor"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/pkg/executor/support"
	"drassi.run/core/util/tar"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type fileHdlCreator func(stack *support.Stack) *command.FileHandler

func testFileNoJob(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Job().Return(nil)

		h := creator(stack)
		err := command.FileRun(t.Context(), h, nil)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testFileNoStep(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Leaf().Return(nil).AnyTimes()
		stack.EXPECT().Root().Return(nil).AnyTimes()
		stack.EXPECT().Stack().Return(nil).AnyTimes()

		h := creator(stack)
		err := command.FileRun(t.Context(), h, nil)
		assert.ErrorIs(t, err, ErrNoStepRunning)
	}
}

var (
	mapContent = `
FOO=bar
ABC=xyz
`
	mapContentMap = map[string]string{
		"FOO": "bar",
		"ABC": "xyz",
	}
)

func TestFileAddPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-job", testFileNoJob(ctrl, FileAddPath))

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader("/fir/path\n/second/path")

		job := mock_executor.NewMockJobExecutor(ctrl)
		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Job().Return(job)
		job.EXPECT().AddPath([]string{"/fir/path", "/second/path"})

		h := FileAddPath(stack)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}

func TestFileSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, func(stack *support.Stack) *command.FileHandler {
		return FileSetEnv(stack, nil)
	}))

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().SetEnv(mapContentMap)

		job := mock_executor.NewMockJobExecutor(ctrl)
		job.EXPECT().SetEnv(mapContentMap)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Stack().Return([]executor.StepExecutor{step})
		stack.EXPECT().Job().Return(job)

		h := FileSetEnv(stack, nil)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}

func TestFileSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSaveState))

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().SaveState(mapContentMap)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Root().Return(step)

		h := FileSaveState(stack)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}

func TestFileSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSetOutput))

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().SetOutput(mapContentMap)

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Leaf().Return(step)

		h := FileSetOutput(stack)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}

func TestCreateStepSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, CreateStepSummary))

	t.Run("success", func(t *testing.T) {
		content := "THIS IS A CreateStepSummary"
		r := strings.NewReader(content)

		step := mock_executor.NewMockStepExecutor(ctrl)
		step.EXPECT().CreateStepSummary(gomock.Any()).Do(func(reader io.Reader) error {
			return xtar.Untar(context.Background(), reader, func(header *tar.Header, r io.Reader) error {
				b, err := io.ReadAll(r)
				assert.NoError(t, err)
				assert.Equal(t, content, string(b))
				return nil
			})
		})

		stack := mock_executor.NewMockStack(ctrl)
		stack.EXPECT().Leaf().Return(step)

		h := CreateStepSummary(stack)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}
