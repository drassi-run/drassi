/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"strings"
	"testing"

	mock_cmdtypes "drassi.run/core/mock/command/cmdtypes"
	"drassi.run/core/pkg/command"
	"drassi.run/core/pkg/command/cmdtypes"
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

	res := mock_cmdtypes.NewMockSupportAddPath(ctrl)
	res.EXPECT().AddPath([]string{"/fir/path", "/second/path"})

	h := FileAddPath[cmdtypes.SupportAddPath]()
	err := command.FileRun[cmdtypes.SupportAddPath](h, t.Context(), res, r)
	assert.NoError(t, err)
}

func TestFileSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdtypes.NewMockSupportSetEnv(ctrl)
		res.EXPECT().SetEnv(mapContentMap)

		h := FileSetEnv[cmdtypes.SupportSetEnv](nil)
		err := command.FileRun[cmdtypes.SupportSetEnv](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSetEnv[cmdtypes.SupportSetEnv](nil)
		err := command.FileRun(h, t.Context(), nil, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	})
}

func TestFileSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdtypes.NewMockSupportSaveState(ctrl)
		res.EXPECT().SaveState(mapContentMap)

		h := FileSaveState[cmdtypes.SupportSaveState]()
		err := command.FileRun[cmdtypes.SupportSaveState](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSaveState[cmdtypes.SupportSaveState]()
		err := command.FileRun[cmdtypes.SupportSaveState](h, t.Context(), nil, r)
		assert.ErrorIs(t, err, ErrInvalidFile)
	})
}

func TestFileSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		r := strings.NewReader(mapContent)

		res := mock_cmdtypes.NewMockSupportSetOutput(ctrl)
		res.EXPECT().SetOutput(mapContentMap)

		h := FileSetOutput[cmdtypes.SupportSetOutput]()
		err := command.FileRun[cmdtypes.SupportSetOutput](h, t.Context(), res, r)
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		r := strings.NewReader("foobar")
		h := FileSetOutput[cmdtypes.SupportSetOutput]()
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
