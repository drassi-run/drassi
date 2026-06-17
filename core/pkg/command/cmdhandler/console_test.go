/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"testing"

	mock_cmdtypes "drassi.run/core/mock/command/cmdtypes"
	mock_secret "drassi.run/core/mock/secret"
	"drassi.run/core/pkg/command"
	"drassi.run/core/pkg/command/cmdtypes"
	"drassi.run/core/pkg/flag"
	"drassi.run/core/pkg/secret"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddSecretMask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		h := AddSecretMask[any](sm)
		cmd := &command.Command{Name: "add-mask", Value: ""}
		err := h.Run(t.Context(), nil, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	})

	t.Run("single-secret", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc"))

		h := AddSecretMask[any](sm)
		cmd := &command.Command{Name: "add-mask", Value: "abc"}
		err := h.Run(t.Context(), nil, cmd)
		assert.NoError(t, err)
	})

	t.Run("multi-secret", func(t *testing.T) {
		sm := mock_secret.NewMockMasker(ctrl)
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc\nxyz\r\nfoo  \r  bar"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("abc"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("xyz"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("foo"))
		sm.EXPECT().AddSecret(secret.NewValueSecret("bar"))

		h := AddSecretMask[any](sm)
		cmd := &command.Command{Name: "add-mask", Value: "abc\nxyz\r\nfoo  \r  bar"}
		err := h.Run(t.Context(), nil, cmd)
		assert.NoError(t, err)
	})
}

type consoleHdlCreator[R any] func() *command.ConsoleHandler[R]

func testInvalidCommand[R any](creator consoleHdlCreator[R]) func(t *testing.T) {
	return func(t *testing.T) {
		h := creator()
		r := new(R)
		cmd := new(command.Command)
		err := h.Run(t.Context(), *r, cmd)
		assert.ErrorIs(t, err, command.ErrInvalidCommand)
	}
}

func TestConsoleAddPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-value", testInvalidCommand(ConsoleAddPath[cmdtypes.SupportAddPath]))

	t.Run("success", func(t *testing.T) {
		res := mock_cmdtypes.NewMockSupportAddPath(ctrl)
		res.EXPECT().AddPath([]string{"foobar"})

		h := ConsoleAddPath[cmdtypes.SupportAddPath]()

		cmd := &command.Command{Name: "add-path", Value: "foobar"}
		err := h.Run(t.Context(), res, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetEnv(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("empty-name", testInvalidCommand(func() *command.ConsoleHandler[cmdtypes.SupportSetEnv] {
		return ConsoleSetEnv[cmdtypes.SupportSetEnv](nil)
	}))

	t.Run("success", func(t *testing.T) {
		res := mock_cmdtypes.NewMockSupportSetEnv(ctrl)
		res.EXPECT().SetEnv(map[string]string{"XXX": "set-env-value"})

		h := ConsoleSetEnv[cmdtypes.SupportSetEnv](nil)

		cmd := &command.Command{Name: "set-env", Params: map[string]string{"name": "XXX"}, Value: "set-env-value"}
		err := h.Run(t.Context(), res, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSetOutput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := func() *command.ConsoleHandler[cmdtypes.SupportSetOutput] {
		return ConsoleSetOutput[cmdtypes.SupportSetOutput](flag.Empty, cmdtypes.Discard[cmdtypes.SupportSetOutput]())
	}

	t.Run("empty-name", testInvalidCommand(creator))

	t.Run("success", func(t *testing.T) {
		res := mock_cmdtypes.NewMockSupportSetOutput(ctrl)
		res.EXPECT().SetOutput(map[string]string{"XXX": "set-output-value"})

		h := creator()

		cmd := &command.Command{Name: "set-output", Params: map[string]string{"name": "XXX"}, Value: "set-output-value"}
		err := h.Run(t.Context(), res, cmd)
		assert.NoError(t, err)
	})
}

func TestConsoleSaveState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	creator := func() *command.ConsoleHandler[cmdtypes.SupportSaveState] {
		return ConsoleSaveState[cmdtypes.SupportSaveState](flag.Empty, cmdtypes.Discard[cmdtypes.SupportSaveState]())
	}

	t.Run("empty-name", testInvalidCommand(creator))

	t.Run("success", func(t *testing.T) {
		res := mock_cmdtypes.NewMockSupportSaveState(ctrl)
		res.EXPECT().SaveState(map[string]string{"XXX": "save-state-value"})

		h := creator()

		cmd := &command.Command{Name: "save-state", Params: map[string]string{"name": "XXX"}, Value: "save-state-value"}
		err := h.Run(t.Context(), res, cmd)
		assert.NoError(t, err)
	})
}
