package base

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type ContainerBs struct {
	interfaces.IContainer
	ExpressionNodeBs
	Params []interfaces.IExpressionNode
}

// AddParameter add node as current container's params
func (c *ContainerBs) AddParameter(node interfaces.IExpressionNode) {
	c.Params = append(c.Params, node)
	node.SetContainer(c)
}

// Parameters return values of all parameter of this ContainerBs node.
// This is read-only, so we will return a slice of value
func (c *ContainerBs) Parameters() []interfaces.IExpressionNode {
	var result []interfaces.IExpressionNode
	for _, p := range c.Params {
		result = append(result, p)
	}
	return result
}

func (c *ContainerBs) GetLevel() (level int) {
	return c.Level
}
