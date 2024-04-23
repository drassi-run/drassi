package base

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

type ContainerBs struct {
	expression.IContainer
	ExpressionNodeBs
	Params []expression.IExpNode
}

// AddParameter add node as current container's params
func (c *ContainerBs) AddParameter(node expression.IExpNode) {
	c.Params = append(c.Params, node)
	node.SetContainer(c)
}

// Parameters return values of all parameter of this ContainerBs node.
// This is read-only, so we will return a slice of value
func (c *ContainerBs) Parameters() []expression.IExpNode {
	var result []expression.IExpNode
	for _, p := range c.Params {
		result = append(result, p)
	}
	return result
}

func (c *ContainerBs) GetLevel() (level int) {
	return c.Level
}
