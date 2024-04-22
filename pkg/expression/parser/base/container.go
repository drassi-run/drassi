package base

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type ContainerBase struct {
	interfaces.IContainer
	ExpressionNodeBase
	Params []interfaces.IExpressionNode
}

// AddParameter add node as current container's params
func (c *ContainerBase) AddParameter(node interfaces.IExpressionNode) {
	c.Params = append(c.Params, node)
	node.SetContainer(c)
}

// Parameters return values of all parameter of this ContainerBase node.
// This is read-only, so we will return a slice of value
func (c *ContainerBase) Parameters() []interfaces.IExpressionNode {
	result := make([]interfaces.IExpressionNode, len(c.Params))
	for _, p := range c.Params {
		result = append(result, p)
	}
	return result
}

func (c *ContainerBase) GetLevel() (level int) {
	return c.Level
}
