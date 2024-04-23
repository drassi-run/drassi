package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type evaluationTraceWriter struct {
	tr interfaces.ITraceWriter
	se interfaces.ISecretMasker
}

func newEvaluationTraceWriter(tr interfaces.ITraceWriter, se interfaces.ISecretMasker) interfaces.ITraceWriter {
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
