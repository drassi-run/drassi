package utilio

import (
	"context"
	"io"
)

func NewContextReader(ctx context.Context, r io.Reader) io.Reader {
	if cr, ok := r.(*contextReader); ok {
		r = cr.r
	}
	return &contextReader{
		ctx: ctx,
		r:   r,
	}
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}

func (cr *contextReader) Read(p []byte) (n int, err error) {
	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	default:
		return cr.r.Read(p)
	}
}

func NewContextWriter(ctx context.Context, w io.Writer) io.Writer {
	if cr, ok := w.(*contextWriter); ok {
		w = cr.w
	}
	return &contextWriter{
		ctx: ctx,
		w:   w,
	}
}

type contextWriter struct {
	ctx context.Context
	w   io.Writer
}

func (cw *contextWriter) Write(p []byte) (n int, err error) {
	select {
	case <-cw.ctx.Done():
		return 0, cw.ctx.Err()
	default:
		return cw.w.Write(p)
	}
}
