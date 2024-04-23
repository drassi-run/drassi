package expression

type IExpNode interface {
	ConvertToExpression() string
	ConvertToRealizedExpression(eCtx IEvaluationContext) string
	SetContainer(c IContainer)
	GetContainer() IContainer
	SetName(name string)
	GetName() string
	// TraceFullyRealized indicates whether the evaluation result should be stored on the context and used when the realized result is traced.
	TraceFullyRealized() bool
	GetLevel() (level int)
	Accept(eCtx IEvaluationContext, visitor IExpNodeVisitor) any
	// Value is only meaningful for Literal node
	Value() any
}
