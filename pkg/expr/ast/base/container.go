package base

import (
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Container struct {
	interfaces.Container
	Node
	Params []interfaces.Node
}

// AddParameter add node as current container's params
func (c *Container) AddParameter(node interfaces.Node) {
	c.Params = append(c.Params, node)
	node.SetCtn(c)
}

// Parameters return values of all parameter of this Container node.
// This is read-only, so we will return a slice of value
func (c *Container) Parameters() []interfaces.Node {
	var result []interfaces.Node
	for _, p := range c.Params {
		result = append(result, p)
	}
	return result
}

func (c *Container) GetLevel() (level int) {
	return c.Level
}
