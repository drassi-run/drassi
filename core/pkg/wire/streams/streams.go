package wire_streams

import (
	"io"

	"drassi.run/core/pkg/executor/logging"
	"drassi.run/core/pkg/sandboxer"
	"go.uber.org/dig"
)

func ProvideTo(scope *dig.Scope) error {
	if err := scope.Provide(processCommand, dig.Name("processCommand")); err != nil {
		return err
	}
	if err := scope.Provide(scanProblem, dig.Name("scanProblem")); err != nil {
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

func newStream(p streamsParams) *sandboxer.Streams {
	return &sandboxer.Streams{
		In:  p.StdIn,
		Out: p.StdOut,
		Err: p.StdErr,
	}
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
	Logger         logging.Logger
	ProcessCommand chainedLineHandler `name:"processCommand"`
	ScanProblem    chainedLineHandler `name:"scanProblem"`
	StdOut         io.Writer          `name:"stdout"`
}

type streamErrParams struct {
	dig.In
	Logger         logging.Logger
	ProcessCommand chainedLineHandler `name:"processCommand"`
	ScanProblem    chainedLineHandler `name:"scanProblem"`
	StdErr         io.Writer          `name:"stderr"`
}

func streamOut(p streamOutParams) io.Writer {
	handler := streamHandler(p.StdOut, p.Logger, []chainedLineHandler{
		p.ProcessCommand,
		p.ScanProblem,
	})
	return logging.NewLineWriter(handler)
}

func streamErr(p streamErrParams) io.Writer {
	handler := streamHandler(p.StdErr, p.Logger, []chainedLineHandler{
		p.ProcessCommand,
		p.ScanProblem,
	})
	return logging.NewLineWriter(handler)
}

func streamHandler(w io.Writer, l logging.Logger, lineHandlers []chainedLineHandler) logging.LineHandler {
	return func(line string) error {
		for _, h := range lineHandlers {
			next, err := h(line)
			if err != nil {
				l.Log(logging.TagError, err.Error())
			}
			if !next {
				return err
			}
		}

		_, err := io.WriteString(w, line)
		return err
	}
}
