package parser

type ITraceWriter interface {
	Info(msg string)
	Verbose(msg string)
}

type evaluationTraceWriter struct {
	tr ITraceWriter
	se ISecretMasker
}

func NewEvaluationTraceWriter(tr ITraceWriter, se ISecretMasker) ITraceWriter {
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
