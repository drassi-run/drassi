package expression

import (
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type IEvaluationContext interface {
	State() any
	Masker() secret_masker.ISecretMasker
	Trace() ITraceWriter
	SetTraceResult(node IExpNode, result IEvaluationResult)
	TryGetTraceResult(node IExpNode) (exist bool, result string)
}
