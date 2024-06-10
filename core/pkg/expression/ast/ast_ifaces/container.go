package ast_ifaces

type Container interface {
	ExprNode
	Parameters() []ExprNode
	AddParameter(node ExprNode)
}
