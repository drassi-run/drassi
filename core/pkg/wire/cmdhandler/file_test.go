package wire_cmdhandler

import (
	"archive/tar"
	"context"
	mock_executor "drassi.run/core/mock/executor"
	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/command"
	"drassi.run/core/util/tar"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"strings"
	"testing"
)

type fileHdlCreator func(executor.Supervisor) *command.FileHandler

func testFileNoJob(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Job().Return(nil)

		h := creator(sup)
		err := command.FileRun(t.Context(), h, nil)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testFileNoStep(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(nil)

		h := creator(sup)
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
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Job().Return(job)
		job.EXPECT().AddPath([]string{"/fir/path", "/second/path"}).Return(nil)

		h := FileAddPath(sup)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}

func TestFileSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSetEnv))

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SetEnv(mapContentMap).Return(nil)

		h := FileSetEnv(sup)
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
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SaveState(mapContentMap).Return(nil)

		h := FileSaveState(sup)
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
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SetOutput(mapContentMap).Return(nil)

		h := FileSetOutput(sup)
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
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().CreateStepSummary(gomock.Any()).Do(func(reader io.Reader) error {
			return xtar.Untar(context.Background(), reader, func(header *tar.Header, r io.Reader) error {
				b, err := io.ReadAll(r)
				assert.NoError(t, err)
				assert.Equal(t, content, string(b))
				return nil
			})
		})

		h := CreateStepSummary(sup)
		err := command.FileRun(t.Context(), h, r)
		assert.NoError(t, err)
	})
}
