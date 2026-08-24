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

// AttachScope return new Handler which attach resource and forward to Sink
func AttachScope[R any](ctx context.Context, scope R, s Sink[R]) Handler {
	return &attachScopeHandler[R]{ctx: ctx, scope: scope, sink: s}
}

type attachScopeHandler[R any] struct {
	ctx   context.Context
	sink  Sink[R]
	scope R
}

func (h *attachScopeHandler[R]) Handle(line string) error {
	return h.sink.Emit(h.ctx, h.scope, line)
}

// DetachScope return new Sink which detach (discard) resource and forward to Handler
func DetachScope[R any](h Handler) Sink[R] {
	return &detachScopeSink[R]{hdl: h}
}

type detachScopeSink[R any] struct {
	hdl Handler
}

func (h *detachScopeSink[R]) Emit(_ context.Context, _ R, line string) error {
	return h.hdl.Handle(line)
}
