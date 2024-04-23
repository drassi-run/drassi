package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type evaluationTraceWriter struct {
	tr expression.ITraceWriter
	se secret_masker.ISecretMasker
}

func newEvaluationTraceWriter(tr expression.ITraceWriter, se secret_masker.ISecretMasker) expression.ITraceWriter {
	if se == nil {
		panic("secret secretMasker must be provider")
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
