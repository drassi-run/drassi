/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
)

type Handler[R any] interface {
	Handle(context.Context, R, string) error
}

type HandlerFunc[R any] func(context.Context, R, string) error

func (f HandlerFunc[R]) Handle(ctx context.Context, r R, line string) error {
	return f(ctx, r, line)
}
