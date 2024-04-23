package expression

type IContainer interface {
	IExpNode
	Parameters() []IExpNode
	AddParameter(node IExpNode)
}
