package chunk

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"time"

	xio "drassi.run/core/util/io"
)

type Option func(*option)

type option struct {
	bufferSize    int
	softChunkSize int
	lineSafe      bool
	followInput   atomic.Bool
}

func WithBufferSize(size int) Option {
	return func(o *option) {
		o.bufferSize = size
	}
}

func WithSoftChunkSize(size int) Option {
	return func(o *option) {
		o.softChunkSize = size
	}
}

// WithLineSafety enable line-safety, so line won't be split across chunks (middle-line cut)
func WithLineSafety(b bool) Option {
	return func(o *option) {
		o.lineSafe = b
	}
}

// WithFollowInput continuous read from Reader until got Complete signal
func WithFollowInput(b bool) Option {
	return func(o *option) {
		o.followInput.Store(b)
	}
}

type Handler = func(context.Context, io.Reader, int64) error

type Processor struct {
	option

	reader     io.Reader
	lineCount  int64
	chunkIndex int64
}

func NewProcessor(r io.Reader, opts ...Option) *Processor {
	p := &Processor{reader: r}
	p.softChunkSize = 2 * 1024 * 1024 // 2MB

	for _, opt := range opts {
		opt(&p.option)
	}
	return p
}

func (p *Processor) Run(ctx context.Context, fn Handler) error {
	if c, ok := p.reader.(io.Closer); ok {
		defer c.Close()
	}

	r := xio.NewContextReader(ctx, p.reader)

	var br *bufio.Reader
	if p.bufferSize > 0 {
		br = bufio.NewReaderSize(r, p.bufferSize)
	} else {
		br = bufio.NewReader(r)
	}

	buf := new(bytes.Buffer)
	for {
		line, err := br.ReadSlice('\n')
		switch {
		case err == nil:
			if _, err = buf.Write(line); err != nil {
				return err
			}
			p.lineCount++
			if buf.Len() >= p.softChunkSize {
				if err = p.process(ctx, buf, fn); err != nil {
					return err
				}
			}
		case errors.Is(err, io.EOF):
			if _, err = buf.Write(line); err != nil {
				return err
			}
			if !p.lineSafe && buf.Len() >= p.softChunkSize {
				if err = p.process(ctx, buf, fn); err != nil {
					return err
				}
			}
			if p.followInput.Load() {
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if buf.Len() > 0 {
				p.lineCount++
				if err = p.process(ctx, buf, fn); err != nil {
					return err
				}
			}
			return nil
		case errors.Is(err, bufio.ErrBufferFull):
			if _, err = buf.Write(line); err != nil {
				return err
			}
			if !p.lineSafe && buf.Len() >= p.softChunkSize {
				if err = p.process(ctx, buf, fn); err != nil {
					return err
				}
			}
		default:
			return err
		}
	}
}

func (p *Processor) process(ctx context.Context, buf *bytes.Buffer, fn Handler) error {
	if err := fn(ctx, buf, p.chunkIndex); err != nil {
		return err
	}

	buf.Reset()
	p.chunkIndex++
	return nil
}

func (p *Processor) Complete() {
	p.followInput.Store(false)
}
