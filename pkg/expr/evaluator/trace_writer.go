package evaluator

import (
	"log/slog"

	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type traceWriter struct {
	tr interfaces.TraceWriter
	sm secret_masker.SecretMasker
}

func newEvaluationTraceWriter(m secret_masker.SecretMasker) interfaces.TraceWriter {
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
