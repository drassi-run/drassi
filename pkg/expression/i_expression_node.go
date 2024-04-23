package expression

type IExpressionNode interface {
	ConvertToExpression() string
	ConvertToRealizedExpression(eCtx IEvaluationContext) string

	// EvaluateCore(eCtx IEvaluationContext) any

	SetContainer(c IContainer)
	GetContainer() IContainer
	GetName() string
	SetName(name string)
	TraceFullyRealized() bool
	GetLevel() (level int)
	Accept(eCtx IEvaluationContext, visitor IExpressionNodeVisitor) any
	// Value is only meaningful for Literal node
	Value() any
}
