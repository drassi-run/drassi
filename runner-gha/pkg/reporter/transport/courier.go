package transport

import (
	"context"
	"io"
	"os"
	"sync"

	"drassi.run/gha-runner/pkg/chunk"
	"drassi.run/gha-runner/pkg/reporter/service"
	"github.com/chainguard-dev/clog"
)

type Courier interface {
	NewFile(ctx context.Context, f string) error
	DoneFile(ctx context.Context, f string) error
	Complete(ctx context.Context, lineCount int64) error
}

// fileCourier uploads the entire file once all writes are completed
type fileCourier struct {
	uploader service.Uploader
	waiter   chunk.Waiter
}

func (c *fileCourier) SetWaiter(waiter chunk.Waiter) {
	c.waiter = waiter
}

func (c *fileCourier) NewFile(context.Context, string) error {
	return nil
}

func (c *fileCourier) DoneFile(ctx context.Context, f string) error {
	if r, err := os.Open(f); err != nil {
		return err
	} else {
		defer r.Close()
		return c.uploader.Upload(ctx, r)
	}
}

func (c *fileCourier) Complete(ctx context.Context, lineCount int64) error {
	return c.uploader.Complete(ctx, lineCount)
}

const chunkSize = 2 * 1024 * 1024 // 2MiB

// chunkCourier follow input file, splits it into chunks and uploads them as soon as they are ready
type chunkCourier struct {
	uploader service.Uploader
	waiter   chunk.Waiter

	ps *chunk.Chunker
	wg sync.WaitGroup
}

func (c *chunkCourier) NewFile(ctx context.Context, f string) error {
	r, err := os.Open(f)
	if err != nil {
		return err
	}

	opts := []chunk.Option{
		chunk.WithSoftLimit(chunkSize),
		chunk.WithLineSafety(true),
		chunk.WithFollowInput(true),
	}
	if c.waiter != nil {
		opts = append(opts, chunk.WithWaiter(c.waiter))
	}
	c.ps = chunk.NewChunker(r, opts...)

	go c.run(ctx)

	return nil
}

func (c *chunkCourier) run(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	handle := func(ctx context.Context, r io.Reader, i int64) error {
		return c.uploader.Upload(ctx, r)
	}

	if err := c.ps.Run(ctx, handle); err != nil {
		clog.Errorf("chunk upload error: %v", err)
		return
	}
}

func (c *chunkCourier) DoneFile(context.Context, string) error {
	if c.ps != nil {
		c.ps.Complete()
	}
	c.wg.Wait()

	return nil
}

func (c *chunkCourier) Complete(ctx context.Context, lineCount int64) error {
	return c.uploader.Complete(ctx, lineCount)
}
