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
	detachResourceHandler bool // use stream.detachResourceHandler as stream.ResourceHandler
}

func UseDetachResourceHandler(b bool) Option {
	return func(o *options) {
		o.detachResourceHandler = b
	}
}

func Module(opts ...Option) *wire.Module {
	o := &options{
		detachResourceHandler: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	fn := func(scope *dig.Scope) error {
		if err := scope.Provide(mdw.ProcessCommand[exec.Milieu], dig.Name("processCommand")); err != nil {
			return fmt.Errorf("provide 'processCommand' stream.Middleware: %w", err)
		}
		if err := scope.Provide(mdw.ScanProblem[exec.Milieu], dig.Name("scanProblem")); err != nil {
			return fmt.Errorf("provide 'scanProblem' stream.Middleware: %w", err)
		}
		if err := scope.Provide(mdw.MaskSecret[exec.Milieu], dig.Name("maskSecret")); err != nil {
			return fmt.Errorf("provide 'maskSecret' stream.Middleware: %w", err)
		}
		if o.detachResourceHandler {
			if err := scope.Provide(stream.NewDetachResourceHandler[exec.Milieu]); err != nil {
				return fmt.Errorf("provide 'detach' stream.ResourceHandler: %w", err)
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

type streamParams[R any] struct {
	dig.In

	Handler        stream.ResourceHandler[R]
	ProcessCommand mdw.Middleware[R] `name:"processCommand"`
	ScanProblem    mdw.Middleware[R] `name:"scanProblem"`
	MaskSecret     mdw.Middleware[R] `name:"maskSecret"`
}

func attachMiddleware[R any](p streamParams[R]) stream.ResourceHandler[R] {
	handler := p.Handler
	middlewares := []mdw.Middleware[R]{p.ProcessCommand, p.ScanProblem, p.MaskSecret}
	for _, mw := range slices.Backward(middlewares) {
		handler = mw(handler)
	}
	return handler
}

func newStreamFactory[R any](h stream.ResourceHandler[R]) stream.Factory[R] {
	return stream.NewFactory[R](
		stream.WithStdout(h),
		stream.WithStderr(h),
	)
}
