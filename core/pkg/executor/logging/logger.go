package logging

import (
	"fmt"
	"io"
	"strings"
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

type Logger interface {
	Logf(tag, format string, a ...any)
	EnableDebug(b bool)
}

func NewLogger(r io.Writer) Logger {
	return &logger{r: r}
}

type logger struct {
	r     io.Writer
	debug bool
}

func (l *logger) EnableDebug(b bool) {
	l.debug = b
}

func (l *logger) Logf(tag, format string, a ...any) {
	if tag == TagDebug && !l.debug {
		return
	}

	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	_, _ = io.WriteString(l.r, message)
}

func Groupf(l Logger, format string, a ...any) func() {
	l.Logf(TagGroup, format, a...)
	return func() {
		l.Logf(TagEndGroup, "")
	}
}

func Logf(l Logger, format string, a ...any) {
	l.Logf("", format, a...)
}

func Debugf(l Logger, format string, a ...any) {
	l.Logf(TagDebug, format, a...)
}

func Noticef(l Logger, format string, a ...any) {
	l.Logf(TagNotice, format, a...)
}

func Errorf(l Logger, format string, a ...any) {
	l.Logf(TagError, format, a...)
}

func Sectionf(l Logger, format string, a ...any) {
	l.Logf(TagSection, format, a...)
}

func Warningf(l Logger, format string, a ...any) {
	l.Logf(TagWarning, format, a...)
}

func Commandf(l Logger, format string, a ...any) {
	l.Logf(TagCommand, format, a...)
}
