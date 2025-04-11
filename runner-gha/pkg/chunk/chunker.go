package chunk

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"drassi.run/core/util/io"
)

type Chunker struct {
	option

	reader     io.Reader
	lineCount  int64
	chunkIndex int64
}

func NewChunker(r io.Reader, opts ...Option) *Chunker {
	p := &Chunker{reader: r}
	p.softLimit = 2 * 1024 * 1024 // 2MB
	p.waiter = DurationWaiter(500 * time.Millisecond)

	for _, opt := range opts {
		opt(&p.option)
	}
	return p
}

type Handler = func(context.Context, io.Reader, int64) error

func (c *Chunker) Run(ctx context.Context, fn Handler) error {
	if cl, ok := c.reader.(io.Closer); ok {
		defer cl.Close()
	}

	r := xio.NewContextReader(ctx, c.reader)

	var br *bufio.Reader
	if c.bufferSize > 0 {
		br = bufio.NewReaderSize(r, c.bufferSize)
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
			c.lineCount++
			if buf.Len() >= c.softLimit {
				if err = c.process(ctx, buf, fn); err != nil {
					return err
				}
			}
		case errors.Is(err, io.EOF):
			if _, err = buf.Write(line); err != nil {
				return err
			}
			if !c.lineSafe && buf.Len() >= c.softLimit {
				if err = c.process(ctx, buf, fn); err != nil {
					return err
				}
			}
			if c.followInput.Load() {
				c.waiter.Wait()
				continue
			}
			if buf.Len() > 0 {
				c.lineCount++
				if err = c.process(ctx, buf, fn); err != nil {
					return err
				}
			}
			return nil
		case errors.Is(err, bufio.ErrBufferFull):
			if _, err = buf.Write(line); err != nil {
				return err
			}
			if !c.lineSafe && buf.Len() >= c.softLimit {
				if err = c.process(ctx, buf, fn); err != nil {
					return err
				}
			}
		default:
			return err
		}
	}
}

func (c *Chunker) process(ctx context.Context, buf *bytes.Buffer, fn Handler) error {
	if err := fn(ctx, buf, c.chunkIndex); err != nil {
		return err
	}

	buf.Reset()
	c.chunkIndex++
	return nil
}

func (c *Chunker) Complete() {
	c.followInput.Store(false)
}
