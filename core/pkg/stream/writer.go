package stream

import (
	"bytes"
	"errors"
	"io"
)

type Handler = func(line string) error

type lineWriter struct {
	closed  bool
	buffer  bytes.Buffer
	handler Handler
}

// NewLineWriter return an [io.Writer] that split input into lines and forward to the [Handler]
func NewLineWriter(handler Handler) io.Writer {
	return &lineWriter{
		handler: handler,
	}
}

func (w *lineWriter) Write(p []byte) (int, error) {
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
		if err = w.handler(w.buffer.String()); err != nil {
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
		return w.handler(s)
	}
	return nil
}
