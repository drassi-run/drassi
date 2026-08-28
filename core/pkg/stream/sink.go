/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import "context"

type Sink[R any] interface {
	Emit(context.Context, R, string) error
}

type SinkFunc[R any] func(context.Context, R, string) error

func (s SinkFunc[R]) Emit(ctx context.Context, res R, line string) error {
	return s(ctx, res, line)
}

// AttachScope return new Handler which attach resource and forward to Sink
func AttachScope[R any](ctx context.Context, res R, sink Sink[R]) Handler {
	return func(line string) error {
		return sink.Emit(ctx, res, line)
	}
}

// DetachScope return new Sink which detach (discard) resource and forward to Handler
func DetachScope[R any](handler Handler) Sink[R] {
	return SinkFunc[R](func(_ context.Context, _ R, line string) error {
		return handler(line)
	})
}
