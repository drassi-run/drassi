package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

type evaluationTraceWriter struct {
	tr expression.ITraceWriter
	se expression.ISecretMasker
}

func newEvaluationTraceWriter(tr expression.ITraceWriter, se expression.ISecretMasker) expression.ITraceWriter {
	if se == nil {
		panic("secret masker must be provider")
	}
	return &evaluationTraceWriter{
		tr: tr,
		se: se,
	}
}

func (e *evaluationTraceWriter) Info(msg string) {
	if e.tr != nil {
		e.tr.Info(e.se.MaskSecrets(msg))
	}
}

func (e *evaluationTraceWriter) Verbose(msg string) {
	if e.tr != nil {
		e.tr.Verbose(e.se.MaskSecrets(msg))
	}
}
