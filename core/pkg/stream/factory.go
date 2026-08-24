/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"io"
)

type Factory[R any] interface {
	Create(ctx context.Context, res R) *Streams
}

type FactoryOption[R any] func(*factory[R])
type SinkFactory[R any] func(name string) Sink[R]

func WithStdin[R any](in io.Reader) FactoryOption[R] {
	return func(f *factory[R]) {
		f.in = in
	}
}

func WithStdout[R any](out SinkFactory[R]) FactoryOption[R] {
	return func(f *factory[R]) {
		f.stdoutCreator = out
	}
}

func WithStderr[R any](err SinkFactory[R]) FactoryOption[R] {
	return func(f *factory[R]) {
		f.stderrCreator = err
	}
}

func NewFactory[R any](opts ...FactoryOption[R]) Factory[R] {
	f := new(factory[R])
	for _, opt := range opts {
		opt(f)
	}
	return f
}

type factory[R any] struct {
	in            io.Reader
	stdoutCreator SinkFactory[R]
	stderrCreator SinkFactory[R]
}

func (f *factory[R]) Create(ctx context.Context, res R) *Streams {
	streams := new(Streams)
	if creator := f.stdoutCreator; creator != nil {
		streams.Out = f.newWriter(ctx, res, creator("stdout"))
	}
	if creator := f.stderrCreator; creator != nil {
		streams.Err = f.newWriter(ctx, res, creator("stderr"))
	}
	streams.In = f.in
	return streams
}

func (f *factory[R]) newWriter(ctx context.Context, res R, sink Sink[R]) io.Writer {
	handler := AttachScope(ctx, res, sink)
	return NewLineWriter(handler)
}
