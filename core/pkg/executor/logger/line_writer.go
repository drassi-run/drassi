package logger

import (
	"bytes"
	"io"
)

type LineHandler = func(line string) error

type lineWriter struct {
	buffer  bytes.Buffer
	handler LineHandler
}

func NewLineWriter(handler LineHandler) io.Writer {
	return &lineWriter{
		handler: handler,
	}
}

func (w *lineWriter) Write(p []byte) (int, error) {
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
