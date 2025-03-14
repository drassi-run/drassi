package wire_streams

import (
	"io"
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
	if err := scope.Provide(streamOut, dig.Name("streamOut")); err != nil {
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

type streamOutParams struct {
	dig.In
	Handler         stream.Handler
	ContextProvider xtypes.ContextProvider
	ProcessCommand  Middleware `name:"processCommand"`
	ScanProblem     Middleware `name:"scanProblem"`
	MaskSecret      Middleware `name:"maskSecret"`
}

func streamOut(p streamOutParams) io.Writer {
	handler := p.Handler
	middlewares := []Middleware{p.ProcessCommand, p.ScanProblem, p.MaskSecret}
	for _, mw := range slices.Backward(middlewares) {
		handler = mw(handler)
	}
	return stream.NewLineWriter(p.ContextProvider, handler)
}

type streamsParams struct {
	dig.In
	StdOut io.Writer `name:"streamOut"`
}

func newStream(p streamsParams) stream.Streams {
	return stream.NewStreams(
		stream.WithStdout(p.StdOut),
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
