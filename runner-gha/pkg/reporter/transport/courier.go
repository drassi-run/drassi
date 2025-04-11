package transport

import (
	"context"
	"io"
	"os"

	"drassi.run/gha-runner/pkg/chunk"
	"drassi.run/gha-runner/pkg/reporter/service"
)

type Courier interface {
	NewFile(ctx context.Context, f string) error
	DoneFile(ctx context.Context, f string) error
	Complete(ctx context.Context, lineCount int64) error
}

// fileCourier uploads the entire file once all writes are completed
type fileCourier struct {
	uploader service.Uploader
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

// chunkCourier splits files into chunks and uploads them as soon as they are ready
type chunkCourier struct {
	uploader service.Uploader
	waiter   chunk.Waiter
}

func (c *chunkCourier) NewFile(ctx context.Context, f string) error {
	r, err := os.Open(f)
	if err != nil {
		return err
	}

	ps := chunk.NewChunker(r,
		chunk.WithSoftLimit(chunkSize),
		chunk.WithLineSafety(true),
		chunk.WithFollowInput(true),
		chunk.WithWaiter(c.waiter),
	)

	handle := func(ctx context.Context, r io.Reader, i int64) error {
		return c.uploader.Upload(ctx, r)
	}

	if err = ps.Run(ctx, handle); err != nil {
		return err
	}
	return nil
}

func (c *chunkCourier) DoneFile(context.Context, string) error {
	return nil
}

func (c *chunkCourier) Complete(ctx context.Context, lineCount int64) error {
	return c.uploader.Complete(ctx, lineCount)
}
