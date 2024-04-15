package parser

type iContainer interface {
	IExpressionNode
	Parameters() []IExpressionNode
	AddParameter(node IExpressionNode)
}

type Container struct {
	iContainer
	ExpressionNode
	params []IExpressionNode
}

// AddParameter add node as current container's params
func (c *Container) AddParameter(node IExpressionNode) {
	c.params = append(c.params, node)
	node.setContainer(c)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (c *Container) Parameters() []IExpressionNode {
	result := make([]IExpressionNode, len(c.params))
	for _, p := range c.params {
		result = append(result, p)
	}
	return result
}

func (c *Container) getLevel() (level int) {
	return c.level
}
