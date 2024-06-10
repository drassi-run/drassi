package secret

import (
	"io"
)

type writer struct {
	masker *Masker
	writer io.Writer
}

func NewWriter(w io.Writer, masker *Masker) io.Writer {
	return &writer{
		masker: masker,
		writer: w,
	}
}

func (w *writer) Write(b []byte) (int, error) {
	s := string(b)
	return w.WriteString(s)
}

func (w *writer) WriteString(s string) (int, error) {
	s = w.masker.Mask(s)
	return io.WriteString(w.writer, s)
}
