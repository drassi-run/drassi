package base

import (
	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type Container struct {
	ast_ifaces.Container
	Node
	Params []ast_ifaces.ExprNode
}

// AddParameter add node as current container's params
func (c *Container) AddParameter(node ast_ifaces.ExprNode) {
	c.Params = append(c.Params, node)
	node.SetCtn(c)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (c *Container) Parameters() []ast_ifaces.ExprNode {
	var result []ast_ifaces.ExprNode
	for _, p := range c.Params {
		result = append(result, p)
	}
	return result
}

func (c *Container) GetLevel() (level int) {
	return c.Level
}
