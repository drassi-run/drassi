/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"io"
)

type Handler interface {
	Handle(context.Context, string) error
}

type HandlerFunc func(context.Context, string) error

func (f HandlerFunc) Handle(ctx context.Context, line string) error {
	return f(ctx, line)
}

func WriteTo(w io.Writer) Handler {
	return HandlerFunc(func(ctx context.Context, msg string) error {
		if _, err := io.WriteString(w, msg); err != nil {
			return err
		}
		if l := len(msg); l == 0 || msg[l-1] != '\n' {
			_, err := w.Write([]byte{'\n'})
			return err
		}
		return nil
	})
}
