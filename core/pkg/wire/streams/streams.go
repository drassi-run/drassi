/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package wire_streams

import (
	"slices"

	"drassi.run/core/pkg/executor"
	"drassi.run/core/pkg/executor/reporter"
	"drassi.run/core/pkg/scribe"
	"drassi.run/core/pkg/stream"
	"drassi.run/core/util/types"
	"go.uber.org/dig"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(ProcessCommand, dig.Name("processCommand")); err != nil {
		return err
	}
	if err := scope.Provide(ScanProblem, dig.Name("scanProblem")); err != nil {
		return err
	}
	if err := scope.Provide(MaskSecret, dig.Name("maskSecret")); err != nil {
		return err
	}

	if err := scope.Provide(streamHandler); err != nil {
		return err
	}
	if err := scope.Provide(newStream, dig.Export(true)); err != nil {
		return err
	}
	if err := scope.Provide(newScribeOutput, dig.Export(true)); err != nil {
		return err
	}

	return nil
}

func streamHandler(rep reporter.Reporter) stream.Handler {
	return stream.HandlerFunc(rep.Log)
}

type streamParams struct {
	dig.In
	Handler         stream.Handler
	ContextProvider xtypes.ContextProvider
	ProcessCommand  Middleware `name:"processCommand"`
	ScanProblem     Middleware `name:"scanProblem"`
	MaskSecret      Middleware `name:"maskSecret"`
}

func newStream(p streamParams) stream.Streams {
	handler := p.Handler
	middlewares := []Middleware{p.ProcessCommand, p.ScanProblem, p.MaskSecret}
	for _, mw := range slices.Backward(middlewares) {
		handler = mw(handler)
	}
	w := stream.NewLineWriter(p.ContextProvider, handler)

	return stream.NewStreams(
		stream.WithStdout(w),
		stream.WithStderr(w),
	)
}

type scribeParams struct {
	dig.In
	Handler    stream.Handler
	MaskSecret Middleware `name:"maskSecret"`
}

func newScribeOutput(p scribeParams) scribe.Output {
	handler := p.Handler
	handler = p.MaskSecret(handler)
	return stream.NewScribeOutput(handler)
}

func Wire(scope *dig.Scope) error {
	return scope.Invoke(registerCallbacks)
}

func registerCallbacks(rep reporter.Reporter, sup executor.Supervisor) error {
	sup.Register(executor.BeforeRunJobCallback(rep.StartJob))
	sup.Register(executor.AfterRunJobCallback(rep.EndJob))
	sup.Register(executor.BeforeRunStepCallback(rep.StartStep))
	sup.Register(executor.AfterRunStepCallback(rep.EndStep))

	return nil
}
