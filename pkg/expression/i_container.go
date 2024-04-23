package expression

type IContainer interface {
	IExpressionNode
	Parameters() []IExpressionNode
	AddParameter(node IExpressionNode)
}
