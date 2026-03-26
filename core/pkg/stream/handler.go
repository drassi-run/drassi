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
	RHandle(context.Context, R, string) error
}

// NewAttachResourceHandler return new Handler which attach resource and forward to ResourceHandler
func NewAttachResourceHandler[R any](ctx context.Context, res R, h ResourceHandler[R]) Handler {
	return &attachResourceHandler[R]{ctx: ctx, res: res, hdl: h}
}

type attachResourceHandler[R any] struct {
	ctx context.Context
	res R
	hdl ResourceHandler[R]
}

func (h *attachResourceHandler[R]) Handle(s string) error {
	return h.hdl.RHandle(h.ctx, h.res, s)
}

// NewDetachResourceHandler return new ResourceHandler which detach (discard) resource and forward to Handler
func NewDetachResourceHandler[R any](h Handler) ResourceHandler[R] {
	return &detachResourceHandler[R]{hdl: h}
}

type detachResourceHandler[R any] struct {
	hdl Handler
}

func (h *detachResourceHandler[R]) RHandle(_ context.Context, _ R, s string) error {
	return h.hdl.Handle(s)
}
