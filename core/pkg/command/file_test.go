/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package command

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"testing"

	mock_sandboxer "drassi.run/core/mock/sandboxer"
	"drassi.run/core/pkg/sandboxer"
	xtar "drassi.run/core/util/tar"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type res string

func (r res) CommandFile(cmd string) string {
	return cmd + string(r)
}

var (
	suffix = res("_suffix")
	layout = &sandboxer.Layout{
		Temp: "/tmp/sandbox",
	}
)

func setupFileCmdMgr(sandbox sandboxer.Sandbox) *fileManager[res] {
	return NewFileManager[res](sandbox).(*fileManager[res])
}

func noopHandler(context.Context, res, io.Reader) error {
	return nil
}

type noopReadCloser struct {
	io.Reader
}

func (n noopReadCloser) Close() error {
	return nil
}

func TestFileManager_Initialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("empty-command", func(tt *testing.T) {
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		mgr := setupFileCmdMgr(sandbox)

		err := mgr.Initialize(ctx, suffix)
		env := mgr.Env("")
		assert.NoError(tt, err)
		assert.Empty(tt, env)
	})

	t.Run("normal", func(tt *testing.T) {
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		sandbox.EXPECT().CopyIn(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		mgr := setupFileCmdMgr(sandbox)
		_ = mgr.Register(NewFileHandler("FIRST", noopHandler))
		_ = mgr.Register(NewFileHandler("SECOND", noopHandler))

		err := mgr.Initialize(ctx, suffix)
		env := mgr.Env(suffix)
		assert.NoError(tt, err)
		assert.EqualValues(tt, env, map[string]string{
			"FIRST":  "/tmp/sandbox/file_commands/FIRST_suffix",
			"SECOND": "/tmp/sandbox/file_commands/SECOND_suffix",
		})
	})
}

func TestFileManager_Process(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	t.Run("empty-command", func(tt *testing.T) {
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		mgr := setupFileCmdMgr(sandbox)
		err := mgr.Process(ctx, suffix)
		assert.NoError(tt, err)
	})

	t.Run("normal", func(tt *testing.T) {
		stringHandler := func(s res) func(context.Context, res, io.Reader) error {
			return func(ctx context.Context, _ res, r io.Reader) error {
				b, err := io.ReadAll(r)
				assert.NoError(tt, err)
				assert.Equal(tt, string(s), string(b))
				return nil
			}
		}

		r1, _ := xtar.ContentReader(map[string]string{"FIRST_suffix": "FIRST file content"})
		r2, _ := xtar.ContentReader(map[string]string{"SECOND_suffix": "SECOND file content"})
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		sandbox.EXPECT().CopyOut(gomock.Any(), "/tmp/sandbox/file_commands/FIRST_suffix").
			Return(&noopReadCloser{r1}, nil)
		sandbox.EXPECT().CopyOut(gomock.Any(), "/tmp/sandbox/file_commands/SECOND_suffix").
			Return(&noopReadCloser{r2}, nil)

		mgr := setupFileCmdMgr(sandbox)
		_ = mgr.Register(NewFileHandler("FIRST", stringHandler("FIRST file content")))
		_ = mgr.Register(NewFileHandler("SECOND", stringHandler("SECOND file content")))

		err := mgr.Process(ctx, suffix)
		assert.NoError(tt, err)
	})

	t.Run("file-not-found", func(tt *testing.T) {
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(nil, fs.ErrNotExist).
			AnyTimes()

		mgr := setupFileCmdMgr(sandbox)
		_ = mgr.Register(NewFileHandler("FIRST", noopHandler))
		_ = mgr.Register(NewFileHandler("SECOND", noopHandler))

		err := mgr.Process(ctx, suffix)
		assert.NoError(tt, err)
	})

	t.Run("copy-error", func(tt *testing.T) {
		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("unexpected error"))

		mgr := setupFileCmdMgr(sandbox)
		_ = mgr.Register(NewFileHandler("FIRST", noopHandler))

		err := mgr.Process(ctx, suffix)
		assert.ErrorContains(tt, err, "unexpected error")
	})

	t.Run("handle-error", func(tt *testing.T) {
		r1, _ := xtar.ContentReader(map[string]string{"FIRST_suffix": "FIRST file content"})

		sandbox := mock_sandboxer.NewMockSandbox(ctrl)
		sandbox.EXPECT().Layout().Return(layout).AnyTimes()
		sandbox.EXPECT().CopyOut(gomock.Any(), gomock.Any()).
			Return(&noopReadCloser{r1}, nil)

		mgr := setupFileCmdMgr(sandbox)
		_ = mgr.Register(NewFileHandler("FIRST", func(_ context.Context, _ res, _ io.Reader) error {
			return errors.New("unexpected error")
		}))

		err := mgr.Process(ctx, suffix)
		assert.ErrorContains(tt, err, "unexpected error")
	})
}
