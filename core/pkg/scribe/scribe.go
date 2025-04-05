/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

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
	ctx   context.Context
	diary Diary
}

func New(ctx context.Context, diary Diary) *Scribe {
	return &Scribe{ctx: ctx, diary: diary}
}

func (s *Scribe) Diary() Diary {
	return s.diary
}

func (s *Scribe) Log(tag, msg string) {
	if tag == TagDebug && !s.diary.EnableDebug() {
		return
	}
	write(s.ctx, s.diary, tag, msg)
}

func (s *Scribe) Writef(format string, args ...any) {
	writef(s.ctx, s.diary, "", format, args...)
}

func (s *Scribe) Debugf(format string, args ...any) {
	if s.diary.EnableDebug() {
		writef(s.ctx, s.diary, TagDebug, format, args...)
	}
}

func (s *Scribe) Noticef(format string, args ...any) {
	writef(s.ctx, s.diary, TagNotice, format, args...)
}

func (s *Scribe) Warningf(format string, args ...any) {
	writef(s.ctx, s.diary, TagWarning, format, args...)
}

func (s *Scribe) Errorf(format string, args ...any) {
	writef(s.ctx, s.diary, TagError, format, args...)
}

func (s *Scribe) Groupf(format string, args ...any) func() {
	writef(s.ctx, s.diary, TagGroup, format, args...)
	return func() {
		writef(s.ctx, s.diary, TagEndGroup, "")
	}
}

func (s *Scribe) Sectionf(format string, args ...any) {
	writef(s.ctx, s.diary, TagSection, format, args...)
}

func (s *Scribe) Commandf(format string, args ...any) {
	writef(s.ctx, s.diary, TagCommand, format, args...)
}

func Log(ctx context.Context, tag, msg string) {
	diary := diaryFromContext(ctx)
	if tag == TagDebug && !diary.EnableDebug() {
		return
	}
	write(ctx, diary, tag, msg)
}

func Writef(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, "", format, args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	if diary.EnableDebug() {
		writef(ctx, diary, TagDebug, format, args...)
	}
}

func Noticef(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagNotice, format, args...)
}

func Warningf(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagWarning, format, args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagError, format, args...)
}

func Groupf(ctx context.Context, format string, args ...any) func() {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagGroup, format, args...)
	return func() {
		writef(ctx, diary, TagEndGroup, "")
	}
}

func Sectionf(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagSection, format, args...)
}

func Commandf(ctx context.Context, format string, args ...any) {
	diary := diaryFromContext(ctx)
	writef(ctx, diary, TagCommand, format, args...)
}

func writef(ctx context.Context, diary Diary, tag, format string, args ...any) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}

	_ = diary.Write(ctx, message)
}

func write(ctx context.Context, diary Diary, tag, message string) {
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}

	_ = diary.Write(ctx, message)
}
