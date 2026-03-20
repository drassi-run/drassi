/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_streams

import (
	"context"
	"slices"

	exec "drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/secret"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	mdw "drassi.run/core/pkg/stream/middleware"
	"go.uber.org/dig"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(mdw.ProcessCommand[exec.Milieu], dig.Name("processCommand")); err != nil {
		return err
	}
	if err := scope.Provide(mdw.ScanProblem[exec.Milieu], dig.Name("scanProblem")); err != nil {
		return err
	}
	if err := scope.Provide(mdw.MaskSecret[exec.Milieu], dig.Name("maskSecret")); err != nil {
		return err
	}

	if err := scope.Provide(newStreamFactory[exec.Milieu], dig.Export(true)); err != nil {
		return err
	}
	if err := scope.Provide(newScribeDiary, dig.Export(true)); err != nil {
		return err
	}

	return nil
}

type streamParams[R any] struct {
	dig.In

	Handler        stream.ResourceHandler[R]
	ProcessCommand mdw.Middleware[R] `name:"processCommand"`
	ScanProblem    mdw.Middleware[R] `name:"scanProblem"`
	MaskSecret     mdw.Middleware[R] `name:"maskSecret"`
}

func newStreamFactory[R any](p streamParams[R]) stream.Factory[R] {
	handler := p.Handler
	middlewares := []mdw.Middleware[R]{p.ProcessCommand, p.ScanProblem, p.MaskSecret}
	for _, mw := range slices.Backward(middlewares) {
		handler = mw(handler)
	}
	return stream.NewFactory[R](
		stream.WithStdout(handler),
		stream.WithStderr(handler),
	)
}

type scribeParams struct {
	dig.In
	Handler      scribe.Handler
	SecretMasker secret.Masker
}

func newScribeDiary(p scribeParams) scribe.Diary {
	h := p.Handler
	sm := p.SecretMasker
	handler := func(ctx context.Context, line string) error {
		line = sm.Mask(line)
		return h(ctx, line)
	}
	return scribe.NewForwardDiary(handler)
}
