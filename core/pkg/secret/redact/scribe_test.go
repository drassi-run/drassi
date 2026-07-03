/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package redact

import (
	"context"
	"errors"
	"testing"

	mock_secret "drassi.run/core/mock/secret"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestScribeHandlerMasksLine(t *testing.T) {
	ctrl := gomock.NewController(t)
	sm := mock_secret.NewMockMasker(ctrl)
	ctx := context.Background()
	handlerErr := errors.New("handler failed")

	sm.EXPECT().Mask("contains secret-token").Return("contains ***")

	handler := ScribeHandler(func(gotCtx context.Context, line string) error {
		assert.Equal(t, ctx, gotCtx)
		assert.Equal(t, "contains ***", line)
		return handlerErr
	}, sm)

	err := handler(ctx, "contains secret-token")
	assert.ErrorIs(t, err, handlerErr)
}
