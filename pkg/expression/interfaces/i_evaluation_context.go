package interfaces

type IEvaluationContext interface {
	State() any
	Masker() ISecretMasker
	Trace() ITraceWriter
	SetTraceResult(node IExpressionNode, result IEvaluationResult)
	TryGetTraceResult(node IExpressionNode) (exist bool, result string)
}
