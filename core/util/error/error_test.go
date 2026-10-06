/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xerror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefine(t *testing.T) {
	t.Run("nil error not refined", func(t *testing.T) {
		var err error
		Refine(&err, "failed to do %s", "task")
		assert.NoError(t, err)
	})

	t.Run("non-nil error refined", func(t *testing.T) {
		err := errors.New("original error")
		Refine(&err, "failed to do %s", "task")
		assert.EqualError(t, err, "failed to do task: original error")
	})
}

func TestRecover(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
		}()
		assert.NoError(t, err)
	})

	t.Run("panic with error", func(t *testing.T) {
		var err error
		origErr := errors.New("boom")
		func() {
			defer Recover(&err)
			panic(origErr)
		}()
		assert.ErrorIs(t, err, origErr)
	})

	t.Run("panic with string", func(t *testing.T) {
		var err error
		func() {
			defer Recover(&err)
			panic("boom")
		}()
		assert.EqualError(t, err, "panic: boom")
	})
}
