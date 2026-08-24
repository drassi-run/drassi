/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_stream

import (
	"fmt"
	"slices"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/stream"
	mdw "drassi.run/core/pkg/stream/middleware"
	"drassi.run/core/wire"
	"go.uber.org/dig"
)

type Option func(o *options)
type options struct {
	detachScopeSink bool // use stream.DetachScope as stream.Sink
}

func UseDetachScopeSink(b bool) Option {
	return func(o *options) {
		o.detachScopeSink = b
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		detachScopeSink: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(mdw.ProcessCommand[exec.Milieu], dig.Name("processCommand")); err != nil {
			return fmt.Errorf("provide 'processCommand' stream.Middleware: %w", err)
		}
		if err := scope.Provide(mdw.DetectProblem[exec.Milieu], dig.Name("detectProblem")); err != nil {
			return fmt.Errorf("provide 'detectProblem' stream.Middleware: %w", err)
		}
		if err := scope.Provide(mdw.MaskSecret[exec.Milieu], dig.Name("maskSecret")); err != nil {
			return fmt.Errorf("provide 'maskSecret' stream.Middleware: %w", err)
		}
		if o.detachScopeSink {
			if err := scope.Provide(stream.DetachScope[exec.Milieu]); err != nil {
				return fmt.Errorf("provide 'detach' stream.Sink: %w", err)
			}
			if err := scope.Provide(staticSinkFactory[exec.Milieu]); err != nil {
				return fmt.Errorf("provide static stream.SinkFactory: %w", err)
			}
		}
		if err := scope.Decorate(attachMiddleware[exec.Milieu]); err != nil {
			return fmt.Errorf("attach stream.Middleware into handler: %w", err)
		}

		if err := scope.Provide(newStreamFactory[exec.Milieu], dig.Export(true)); err != nil {
			return fmt.Errorf("provide stream.Factory: %w", err)
		}

		return nil
	}
	return wire.NewModule("core/stream", fn)
}

func staticSinkFactory[R any](h stream.Sink[R]) stream.SinkFactory[R] {
	return func(name string) stream.Sink[R] { return h }
}

type streamParams[R any] struct {
	dig.In

	SinkFactory    stream.SinkFactory[R]
	ProcessCommand mdw.Middleware[R] `name:"processCommand"`
	DetectProblem  mdw.Middleware[R] `name:"detectProblem"`
	MaskSecret     mdw.Middleware[R] `name:"maskSecret"`
}

func attachMiddleware[R any](p streamParams[R]) stream.SinkFactory[R] {
	middlewares := []mdw.Middleware[R]{p.ProcessCommand, p.DetectProblem, p.MaskSecret}
	return func(name string) stream.Sink[R] {
		s := p.SinkFactory(name)
		for _, mw := range slices.Backward(middlewares) {
			s = mw(s)
		}
		return s
	}
}

func newStreamFactory[R any](f stream.SinkFactory[R]) stream.Factory[R] {
	return stream.NewFactory[R](
		stream.WithStdout(f),
		stream.WithStderr(f),
	)
}
