/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"strings"
	"testing"

	mock_cmdhandler "drassi.run/core/mock/executor/command/cmdhandler"
	"drassi.run/core/pkg/executor/command"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

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

	r := strings.NewReader("/fir/path\n/second/path")

	res := mock_cmdhandler.NewMockSupportAddPath(ctrl)
	res.EXPECT().AddPath([]string{"/fir/path", "/second/path"})

	h := FileAddPath[SupportAddPath]()
	err := command.FileRun[SupportAddPath](h, t.Context(), res, r)
	assert.NoError(t, err)
}

func TestFileSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdhandler.NewMockSupportSetEnv(ctrl)
		res.EXPECT().SetEnv(mapContentMap)

		h := FileSetEnv[SupportSetEnv](nil)
		err := command.FileRun[SupportSetEnv](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSetEnv[SupportSetEnv](nil)
		err := command.FileRun(h, t.Context(), nil, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	})
}

func TestFileSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdhandler.NewMockSupportSaveState(ctrl)
		res.EXPECT().SaveState(mapContentMap)

		h := FileSaveState[SupportSaveState]()
		err := command.FileRun[SupportSaveState](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSaveState[SupportSaveState]()
		err := command.FileRun[SupportSaveState](h, t.Context(), nil, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	})
}

func TestFileSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdhandler.NewMockSupportSetOutput(ctrl)
		res.EXPECT().SetOutput(mapContentMap)

		h := FileSetOutput[SupportSetOutput]()
		err := command.FileRun[SupportSetOutput](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSetOutput[SupportSetOutput]()
		err := command.FileRun(h, t.Context(), nil, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	})
}

func TestCreateStepSummary(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		content := "THIS IS A CreateStepSummary"
		r := strings.NewReader(content)

		h := CreateStepSummary[any]()
		err := command.FileRun(h, t.Context(), nil, r)
		assert.NoError(t, err)
	})
}
