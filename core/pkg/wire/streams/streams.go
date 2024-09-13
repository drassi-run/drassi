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

	return nil
}

type streamsParams struct {
	In  io.Reader `name:"stdin"`
	Out io.Writer `name:"streamOut"`
	Err io.Writer `name:"streamErr"`
}

func newStream(params streamsParams) *sandboxer.Streams {
	return &sandboxer.Streams{
		In:  params.In,
		Out: params.Out,
		Err: params.Err,
	}
}

type streamIOParams struct {
	dig.In
	Logger         logging.Logger
	ProcessCommand chainedLineHandler `name:"processCommand"`
	ScanProblem    chainedLineHandler `name:"scanProblem"`
}

type streamOutParams struct {
	streamIOParams
	Out io.Writer `name:"stdout"`
}

type streamErrParams struct {
	streamIOParams
	Out io.Writer `name:"stdout"`
}

func streamOut(params streamOutParams) io.Writer {
	handler := streamHandler(params.Out, params.Logger, []chainedLineHandler{
		params.ProcessCommand,
		params.ScanProblem,
	})
	return logging.NewLineWriter(handler)
}

func streamErr(params streamErrParams) io.Writer {
	handler := streamHandler(params.Out, params.Logger, []chainedLineHandler{
		params.ProcessCommand,
		params.ScanProblem,
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
