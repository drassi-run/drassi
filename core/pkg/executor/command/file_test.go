package command

import (
	"context"
	mock_sandboxer "drassi.run/core/pkg/sandboxer/mock"
	"errors"
	"go.uber.org/mock/gomock"
	"gotest.tools/v3/assert"
	"io"
	"io/fs"
	"strings"
	"testing"
)

func setupFileCmdMgr() *fileCommandManager {
	return NewFileCommandManager("_suffix").(*fileCommandManager)
}

func noopHandler(r io.Reader) error {
	return nil
}

type noopReadCloser struct {
	io.Reader
}

func (n noopReadCloser) Close() error {
	return nil
}

func TestFileCommandManager_Initialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("empty-command", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		env, err := mgr.Initialize(ctx, sandbox)
		assert.NilError(tt, err)
		assert.Assert(tt, env == nil)
	})

	t.Run("normal", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		_ = mgr.RegisterCommand("FIRST", noopHandler)
		_ = mgr.RegisterCommand("SECOND", noopHandler)

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().GetTempDir().Return("/tmp/sandbox")
		sandbox.EXPECT().CopyIn(gomock.Any(), gomock.Any(), "/tmp/sandbox/file_commands").Return(nil)

		env, err := mgr.Initialize(ctx, sandbox)
		assert.NilError(tt, err)
		assert.DeepEqual(tt, env, map[string]string{
			"FIRST":  "/tmp/sandbox/file_commands/FIRST_suffix",
			"SECOND": "/tmp/sandbox/file_commands/SECOND_suffix",
		})
	})
}

func TestFileCommandManager_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("empty-command", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		err := mgr.Process(ctx, sandbox)
		assert.NilError(tt, err)
	})

	t.Run("normal", func(tt *testing.T) {
		stringHandler := func(t *testing.T, s string) FileCommandHandler {
			return func(r io.Reader) error {
				b, err := io.ReadAll(r)
				assert.NilError(t, err)
				assert.Equal(t, s, string(b))
				return nil
			}
		}

		mgr := setupFileCmdMgr()
		_ = mgr.RegisterCommand("FIRST", stringHandler(tt, "FIRST file content"))
		_ = mgr.RegisterCommand("SECOND", stringHandler(tt, "SECOND file content"))

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().GetTempDir().Return("/tmp/sandbox")
		sandbox.EXPECT().CopyOut(gomock.Any(), "/tmp/sandbox/file_commands/FIRST_suffix").
			Return(&noopReadCloser{strings.NewReader("FIRST file content")}, nil)
		sandbox.EXPECT().CopyOut(gomock.Any(), "/tmp/sandbox/file_commands/SECOND_suffix").
			Return(&noopReadCloser{strings.NewReader("SECOND file content")}, nil)

		err := mgr.Process(ctx, sandbox)
		assert.NilError(tt, err)
	})

	t.Run("file-not-found", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		_ = mgr.RegisterCommand("FIRST", noopHandler)
		_ = mgr.RegisterCommand("SECOND", noopHandler)

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().GetTempDir().Return("/tmp/sandbox")
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(nil, fs.ErrNotExist).
			AnyTimes()

		err := mgr.Process(ctx, sandbox)
		assert.NilError(tt, err)
	})

	t.Run("copy-error", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		_ = mgr.RegisterCommand("FIRST", noopHandler)

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().GetTempDir().Return("/tmp/sandbox")
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("unexpected error"))

		err := mgr.Process(ctx, sandbox)
		assert.Error(tt, err, "unexpected error")
	})

	t.Run("handle-error", func(tt *testing.T) {
		mgr := setupFileCmdMgr()
		_ = mgr.RegisterCommand("FIRST", func(r io.Reader) error {
			return errors.New("unexpected error")
		})

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().GetTempDir().Return("/tmp/sandbox")
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(&noopReadCloser{strings.NewReader("FIRST file content")}, nil)

		err := mgr.Process(ctx, sandbox)
		assert.Error(tt, err, "unexpected error")
	})
}
