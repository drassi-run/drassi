package wire_streams

import (
	"io"
	"slices"

	"drassi.run/core/pkg/executor/logging"
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

	if err := scope.Provide(streamOut, dig.Name("streamOut")); err != nil {
		return err
	}
	if err := scope.Provide(streamErr, dig.Name("streamErr")); err != nil {
		return err
	}

	if err := scope.Provide(newStream, dig.Export(true)); err != nil {
		return err
	}
	if err := scope.Provide(newLog, dig.Export(true)); err != nil {
		return err
	}

	return nil
}

type streamsParams struct {
	dig.In
	StdIn  io.Reader `name:"stdin"`
	StdOut io.Writer `name:"streamOut"`
	StdErr io.Writer `name:"streamErr"`
}

func newStream(p streamsParams) stream.Streams {
	return stream.NewStreams(
		stream.WithStdin(p.StdIn),
		stream.WithStdout(p.StdOut),
		stream.WithStderr(p.StdErr),
	)
}

type logParams struct {
	dig.In
	StdOut io.Writer `name:"stdout"`
}

func newLog(p logParams) logging.Logger {
	return logging.NewLogger(p.StdOut)
}

type streamOutParams struct {
	dig.In
	ContextProvider xtypes.ContextProvider
	ProcessCommand  Middleware `name:"processCommand"`
	ScanProblem     Middleware `name:"scanProblem"`
	MaskSecret      Middleware `name:"maskSecret"`
	StdOut          io.Writer  `name:"stdout"`
}

type streamErrParams struct {
	dig.In
	ContextProvider xtypes.ContextProvider
	ProcessCommand  Middleware `name:"processCommand"`
	ScanProblem     Middleware `name:"scanProblem"`
	MaskSecret      Middleware `name:"maskSecret"`
	StdErr          io.Writer  `name:"stderr"`
}

func streamOut(p streamOutParams) io.Writer {
	handler := streamHandler(p.StdOut, []Middleware{
		p.ProcessCommand,
		p.ScanProblem,
		p.MaskSecret,
	})
	return stream.NewLineWriter(p.ContextProvider, handler)
}

func streamErr(p streamErrParams) io.Writer {
	handler := streamHandler(p.StdErr, []Middleware{
		p.ProcessCommand,
		p.ScanProblem,
		p.MaskSecret,
	})
	return stream.NewLineWriter(p.ContextProvider, handler)
}

func streamHandler(w io.Writer, middlewares []Middleware) stream.Handler {
	h := stream.WriteTo(w)
	for _, mw := range slices.Backward(middlewares) {
		h = mw(h)
	}
	return h
}
