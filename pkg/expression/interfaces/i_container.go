package interfaces

type IContainer interface {
	IExpressionNode
	Parameters() []IExpressionNode
	AddParameter(node IExpressionNode)
}
