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
	"testing"
)

type fileHdlCreator func(executor.Supervisor) *command.FileHandler

func testInvalidFile(ctrl *gomock.Controller, creator fileHdlCreator, r io.Reader) func(t *testing.T) {
	return func(t *testing.T) {
		job := mock_executor.NewMockJobExecutor(ctrl)
		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background()).AnyTimes()
		sup.EXPECT().Job().Return(job).AnyTimes()
		sup.EXPECT().CurrentStep().Return(step).AnyTimes()

		h := creator(sup)
		err := command.FileRun(h, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	}
}

func testFileNoJob(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background()).AnyTimes()
		sup.EXPECT().Job().Return(nil)

		h := creator(sup)
		err := command.FileRun(h, nil)
		assert.ErrorIs(t, err, ErrNoJobRunning)
	}
}

func testFileNoStep(ctrl *gomock.Controller, creator fileHdlCreator) func(t *testing.T) {
	return func(t *testing.T) {
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background()).AnyTimes()
		sup.EXPECT().CurrentStep().Return(nil)

		h := creator(sup)
		err := command.FileRun(h, nil)
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

	invalidFile, _ := xtar.ContentReader(map[string]string{
		"A_FILE": "FOOBAR=hello",
	})
	t.Run("invalid-file", testInvalidFile(ctrl, FileAddPath, invalidFile))

	t.Run("success", func(t *testing.T) {
		r, _ := xtar.ContentReader(map[string]string{
			"": "/fir/path\n/second/path",
		})

		job := mock_executor.NewMockJobExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background())
		sup.EXPECT().Job().Return(job)
		job.EXPECT().AddPath([]string{"/fir/path", "/second/path"}).Return(nil)

		h := FileAddPath(sup)
		err := command.FileRun(h, r)
		assert.NoError(t, err)
	})
}

func TestFileSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSetEnv))

	invalidFile, _ := xtar.ContentReader(map[string]string{
		"A_FILE": "FOOBAR=hello",
	})
	t.Run("invalid-file", testInvalidFile(ctrl, FileSetEnv, invalidFile))

	t.Run("success", func(t *testing.T) {
		r, _ := xtar.ContentReader(map[string]string{
			"": mapContent,
		})

		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background())
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SetEnv(mapContentMap).Return(nil)

		h := FileSetEnv(sup)
		err := command.FileRun(h, r)
		assert.NoError(t, err)
	})
}

func TestFileSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSaveState))

	invalidFile, _ := xtar.ContentReader(map[string]string{
		"A_FILE": "FOOBAR=hello",
	})
	t.Run("invalid-file", testInvalidFile(ctrl, FileSaveState, invalidFile))

	t.Run("success", func(t *testing.T) {
		r, _ := xtar.ContentReader(map[string]string{
			"": mapContent,
		})

		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background())
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SaveState(mapContentMap).Return(nil)

		h := FileSaveState(sup)
		err := command.FileRun(h, r)
		assert.NoError(t, err)
	})
}

func TestFileSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, FileSetOutput))

	invalidFile, _ := xtar.ContentReader(map[string]string{
		"A_FILE": "FOOBAR=hello",
	})
	t.Run("invalid-file", testInvalidFile(ctrl, FileSetOutput, invalidFile))

	t.Run("success", func(t *testing.T) {
		r, _ := xtar.ContentReader(map[string]string{
			"": mapContent,
		})

		step := mock_executor.NewMockStepExecutor(ctrl)
		sup := mock_executor.NewMockSupervisor(ctrl)
		sup.EXPECT().Context().Return(context.Background())
		sup.EXPECT().CurrentStep().Return(step)
		step.EXPECT().SetOutput(mapContentMap).Return(nil)

		h := FileSetOutput(sup)
		err := command.FileRun(h, r)
		assert.NoError(t, err)
	})
}

func TestCreateStepSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("no-step", testFileNoStep(ctrl, CreateStepSummary))

	t.Run("success", func(t *testing.T) {
		content := "THIS IS A CreateStepSummary"
		r, _ := xtar.ContentReader(map[string]string{
			"": content,
		})

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
		err := command.FileRun(h, r)
		assert.NoError(t, err)
	})
}
