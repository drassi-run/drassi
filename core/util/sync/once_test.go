/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xsync

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleton(t *testing.T) {
	t.Run("executes once", func(t *testing.T) {
		var calls atomic.Int32
		fn := Singleton(func(ctx context.Context) (string, error) {
			calls.Add(1)
			return "hello", nil
		})

		ctx := t.Context()
		val1, err1 := fn(ctx)
		require.NoError(t, err1)
		require.Equal(t, "hello", val1)

		val2, err2 := fn(ctx)
		require.NoError(t, err2)
		require.Equal(t, "hello", val2)

		require.Equal(t, int32(1), calls.Load())
	})

	t.Run("concurrent calls invoke once", func(t *testing.T) {
		var calls atomic.Int32
		fn := Singleton(func(ctx context.Context) (int, error) {
			calls.Add(1)
			return 42, nil
		})

		var wg sync.WaitGroup
		ctx := t.Context()
		for range 50 {
			wg.Go(func() {
				val, err := fn(ctx)
				require.NoError(t, err)
				require.Equal(t, 42, val)
			})
		}
		wg.Wait()

		require.Equal(t, int32(1), calls.Load())
	})

	t.Run("error memoized", func(t *testing.T) {
		expectedErr := errors.New("boom")
		var calls atomic.Int32
		fn := Singleton(func(ctx context.Context) (string, error) {
			calls.Add(1)
			return "", expectedErr
		})

		ctx := t.Context()
		_, err1 := fn(ctx)
		require.ErrorIs(t, err1, expectedErr)

		_, err2 := fn(ctx)
		require.ErrorIs(t, err2, expectedErr)

		require.Equal(t, int32(1), calls.Load())
	})

	t.Run("panic recovery", func(t *testing.T) {
		var calls atomic.Int32
		fn := Singleton(func(ctx context.Context) (string, error) {
			calls.Add(1)
			panic("disaster")
		})

		ctx := t.Context()
		require.PanicsWithValue(t, "disaster", func() {
			_, _ = fn(ctx)
		})

		require.PanicsWithValue(t, "disaster", func() {
			_, _ = fn(ctx)
		})

		require.Equal(t, int32(1), calls.Load())
	})
}
