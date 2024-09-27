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
	Log(tag, format string, a ...any)
}

func NewLogger(r io.Writer) Logger {
	return &logger{r: r}
}

type logger struct {
	r io.Writer
}

func (l *logger) Log(tag, format string, a ...any) {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	if tag != "" {
		message = fmt.Sprintf("##[%s]%s", tag, message)
	}
	if !strings.HasSuffix(message, "\n") {
		message = message + "\n"
	}
	_, _ = io.WriteString(l.r, message)
}
