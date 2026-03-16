/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
)

type Handler interface {
	Handle(string) error
}

type ResourceHandler[R any] interface {
	Handle(context.Context, R, string) error
}

// NewHandlerWithResource return new Handler that attach resource to it
func NewHandlerWithResource[R any](ctx context.Context, res R, h ResourceHandler[R]) Handler {
	return &handlerWithResource[R]{ctx: ctx, res: res, hdl: h}
}

type handlerWithResource[R any] struct {
	ctx context.Context
	res R
	hdl ResourceHandler[R]
}

func (h *handlerWithResource[R]) Handle(s string) error {
	return h.hdl.Handle(h.ctx, h.res, s)
}
