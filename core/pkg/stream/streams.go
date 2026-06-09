/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"errors"
	"io"
)

type Streams struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

func (s *Streams) Close() error {
	errs := make([]error, 0, 3)
	if c, ok := s.In.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	if c, ok := s.Out.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	if c, ok := s.Err.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

type Factory[R any] interface {
	Create(ctx context.Context, res R) *Streams
}

func NewFactory[R any](opts ...FactoryOption[R]) Factory[R] {
	f := new(factory[R])
	for _, opt := range opts {
		opt(f)
	}
	return f
}

type factory[R any] struct {
	in         io.Reader
	outHandler ResourceHandler[R]
	errHandler ResourceHandler[R]
}

func (f *factory[R]) Create(ctx context.Context, res R) *Streams {
	out := f.newWriter(ctx, res, f.outHandler)
	err := f.newWriter(ctx, res, f.errHandler)
	return &Streams{
		In:  f.in,
		Out: out,
		Err: err,
	}
}

func (f *factory[R]) newWriter(ctx context.Context, res R, hdl ResourceHandler[R]) io.Writer {
	handler := NewAttachResourceHandler(ctx, res, hdl)
	return NewLineWriter(handler)
}

type FactoryOption[R any] func(*factory[R])

func WithStdin[R any](in io.Reader) FactoryOption[R] {
	return func(f *factory[R]) {
		f.in = in
	}
}

func WithStdout[R any](out ResourceHandler[R]) FactoryOption[R] {
	return func(f *factory[R]) {
		f.outHandler = out
	}
}

func WithStderr[R any](err ResourceHandler[R]) FactoryOption[R] {
	return func(f *factory[R]) {
		f.errHandler = err
	}
}
