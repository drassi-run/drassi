package evaluator

import (
	"log/slog"

	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/secret_masker"
)

type traceWriter struct {
	tr ast_ifaces.TraceWriter
	sm secret_masker.Interface
}

func newEvaluationTraceWriter(m secret_masker.Interface) ast_ifaces.TraceWriter {
	l := logrus.New()
	return &traceWriter{
		tr: slog.New(sloglogrus.Option{Level: slog.LevelDebug, Logger: l}.NewLogrusHandler()),
		sm: m,
	}
}

func (e *traceWriter) Info(msg string, args ...any) {
	if e.tr != nil {
		e.tr.Info(e.sm.Mask(msg), args...)
	}
}

func (e *traceWriter) Debug(msg string, args ...any) {
	if e.tr != nil {
		e.tr.Debug(e.sm.Mask(msg), args...)
	}
}
