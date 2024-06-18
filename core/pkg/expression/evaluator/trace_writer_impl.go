package evaluator

import (
	"log/slog"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/secret_masker"
	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"
)

type traceWriter struct {
	tr ast_ifaces.TraceWriter
	sm secret_masker.Interface
}

func newEvaluationTraceWriter(m secret_masker.Interface) ast_ifaces.TraceWriter {
	if m == nil {
		panic("secret masker must be provided")
	}
	l := logrus.New()
	return &traceWriter{
		tr: slog.New(sloglogrus.Option{Level: slog.LevelDebug, Logger: l}.NewLogrusHandler()),
		sm: m,
	}
}

func (e *traceWriter) Info(msg string, args ...any) {
	if e.tr != nil {
		e.tr.Info(e.sm.MaskSecrets(msg), args...)
	}
}

func (e *traceWriter) Debug(msg string, args ...any) {
	if e.tr != nil {
		e.tr.Debug(e.sm.MaskSecrets(msg), args...)
	}
}
