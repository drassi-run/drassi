/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"bytes"
	"errors"
	"io"

	"drassi.run/core/util/types"
)

type lineWriter struct {
	closed     bool
	buffer     bytes.Buffer
	handler    Handler
	contextual xtypes.ContextProvider
}

// NewLineWriter return an [io.Writer] that split input into lines and forward to the [Handler]
func NewLineWriter(c xtypes.ContextProvider, h Handler) io.Writer {
	return &lineWriter{
		handler:    h,
		contextual: c,
	}
}

func (w *lineWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, errors.New("attempt to write to closed writer")
	}

	ctx := w.contextual.Context()
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
		if err = w.handler.Handle(ctx, w.buffer.String()); err != nil {
			return written, err
		}
		w.buffer.Reset()
	}
	return written, nil
}

func (w *lineWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *lineWriter) Close() error {
	w.closed = true
	defer w.buffer.Reset()
	if s := w.buffer.String(); s != "" {
		ctx := w.contextual.Context()
		return w.handler.Handle(ctx, s)
	}
	return nil
}
