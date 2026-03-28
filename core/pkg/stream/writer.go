/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"bytes"
	"context"
	"errors"
	"io"
)

type lineWriter[R any] struct {
	ctx     context.Context
	res     R
	handler Handler[R]
	buffer  bytes.Buffer
	closed  bool
}

// NewLineWriter return an [io.Writer] that split input into lines and forward to the [Handler]
func NewLineWriter[R any](ctx context.Context, res R, h Handler[R]) io.Writer {
	return &lineWriter[R]{
		ctx:     ctx,
		res:     res,
		handler: h,
	}
}

func (w *lineWriter[R]) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errors.New("attempt to write to closed writer")
	}

	buf := bytes.NewBuffer(p)
	written := 0
	for {
		line, err := buf.ReadBytes('\n')
		n, _ := w.buffer.Write(line)
		written += n
		if err != nil {
			if err == io.EOF {
				break
			}
			return written, err
		}
		if err = w.handler.Handle(w.ctx, w.res, w.buffer.String()); err != nil {
			return written, err
		}
		w.buffer.Reset()
	}
	return written, nil
}

func (w *lineWriter[R]) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *lineWriter[R]) Close() error {
	w.closed = true
	defer w.buffer.Reset()
	if s := w.buffer.String(); s != "" {
		return w.handler.Handle(w.ctx, w.res, s)
	}
	return nil
}
