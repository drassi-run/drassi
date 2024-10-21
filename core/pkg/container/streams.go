package container

import "io"

type Streams interface {
	In() io.Reader
	Out() io.Writer
	Err() io.Writer
}

type StreamsOption func(*streams)

type streams struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func (s *streams) In() io.Reader {
	return s.in
}

func (s *streams) Out() io.Writer {
	return s.out
}

func (s *streams) Err() io.Writer {
	return s.err
}

func NewStreams(opts ...StreamsOption) Streams {
	s := new(streams)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithStdin(in io.Reader) StreamsOption {
	return func(s *streams) {
		s.in = in
	}
}

func WithStdout(out io.Writer) StreamsOption {
	return func(s *streams) {
		s.out = out
	}
}

func WithStderr(err io.Writer) StreamsOption {
	return func(s *streams) {
		s.err = err
	}
}
