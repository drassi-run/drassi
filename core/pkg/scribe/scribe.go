package scribe

import (
	"context"
	"fmt"
)

const (
	TagGroup    = "group"
	TagEndGroup = "endgroup"
	TagSection  = "section"
	TagCommand  = "command"
	TagError    = "error"
	TagWarning  = "warning"
	TagNotice   = "notice"
	TagDebug    = "debug"
)

type Scribe struct {
	ctx context.Context
	out Output
}

func New(ctx context.Context, out Output) *Scribe {
	return &Scribe{ctx: ctx, out: out}
}

func (s *Scribe) Handler() Output {
	return s.out
}

func (s *Scribe) Log(tag, msg string) {
	if tag == TagDebug && !s.out.EnableDebug() {
		return
	}
	write(s.ctx, s.out, tag, msg)
}

func (s *Scribe) Writef(format string, args ...any) {
	writef(s.ctx, s.out, "", format, args...)
}

func (s *Scribe) Debugf(format string, args ...any) {
	if s.out.EnableDebug() {
		writef(s.ctx, s.out, TagDebug, format, args...)
	}
}

func (s *Scribe) Noticef(format string, args ...any) {
	writef(s.ctx, s.out, TagNotice, format, args...)
}

func (s *Scribe) Warningf(format string, args ...any) {
	writef(s.ctx, s.out, TagWarning, format, args...)
}

func (s *Scribe) Errorf(format string, args ...any) {
	writef(s.ctx, s.out, TagError, format, args...)
}

func (s *Scribe) Groupf(format string, args ...any) func() {
	writef(s.ctx, s.out, TagGroup, format, args...)
	return func() {
		writef(s.ctx, s.out, TagEndGroup, "")
	}
}

func (s *Scribe) Sectionf(format string, args ...any) {
	writef(s.ctx, s.out, TagSection, format, args...)
}

func (s *Scribe) Commandf(format string, args ...any) {
	writef(s.ctx, s.out, TagCommand, format, args...)
}

func Log(ctx context.Context, tag, msg string) {
	handler := handlerFromContext(ctx)
	if tag == TagDebug && !handler.EnableDebug() {
		return
	}
	write(ctx, handler, tag, msg)
}

func Writef(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, "", format, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	if handler.EnableDebug() {
		writef(ctx, handler, TagDebug, format, args...)
	}
}

func Noticef(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagNotice, format, args...)
}

func Warningf(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagWarning, format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagError, format, args...)
}

func Groupf(ctx context.Context, format string, args ...any) func() {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagGroup, format, args...)
	return func() {
		writef(ctx, handler, TagEndGroup, "")
	}
}

func Sectionf(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagSection, format, args...)
}

func Commandf(ctx context.Context, format string, args ...any) {
	handler := handlerFromContext(ctx)
	writef(ctx, handler, TagCommand, format, args...)
}

func writef(ctx context.Context, output Output, tag, format string, args ...any) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}

	_ = output.Inscribe(ctx, message)
}

func write(ctx context.Context, output Output, tag, message string) {
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}

	_ = output.Inscribe(ctx, message)
}
