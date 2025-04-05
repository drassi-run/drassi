package reporter

import (
	"io"
	"os"
	"path/filepath"
)

type Output struct {
	dir string
	w   io.WriteCloser
}

func (o *Output) Write(p []byte) (n int, err error) {
	if o.w == nil {
		return 0, nil
	}
	return o.w.Write(p)
}

func (o *Output) Close() error {
	if o.w != nil {
		return o.w.Close()
	}
	return nil
}

func (o *Output) Next(file string) (string, error) {
	if err := o.Close(); err != nil {
		return "", err
	}
	p := filepath.Join(o.dir, file)
	if f, err := os.Create(p); err != nil {
		return "", err
	} else {
		o.w = f
		return p, nil
	}
}
