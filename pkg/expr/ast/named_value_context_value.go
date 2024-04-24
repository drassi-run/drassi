package ast

import (
	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type ContextValueNode struct {
	NamedValue
	base.Node
}

func (c *ContextValueNode) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitContextValueNode(eCtx, c)
}

func (c *ContextValueNode) Value() any {
	panic("not implemented")
}

func (c *ContextValueNode) TraceFullyRealized() bool {
	return true
}

func (c *ContextValueNode) GetCtn() interfaces.Container {
	return c.Ctn
}

func (c *ContextValueNode) SetCtn(cc interfaces.Container) {
	c.Ctn = cc
}

func (c *ContextValueNode) SetName(name string) {
	c.Node.Name = name
}

func (c *ContextValueNode) GetName() string {
	return c.Name
}
