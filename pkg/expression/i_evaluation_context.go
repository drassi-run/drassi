package expression

import (
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type IEvaluationContext interface {
	State() any
	Masker() interfaces.ISecretMasker
	Trace() ITraceWriter
	SetTraceResult(node IExpressionNode, result IEvaluationResult)
	TryGetTraceResult(node IExpressionNode) (exist bool, result string)
}
