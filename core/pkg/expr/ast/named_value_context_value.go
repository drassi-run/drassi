package ast

import (
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/base"
)

type ContextValueNode struct {
	NamedValue
	base.Node
}

func (c *ContextValueNode) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitContextValueNode(eCtx, c)
}

func (c *ContextValueNode) Value() any {
	panic("not implemented")
}

func (c *ContextValueNode) TraceFullyRealized() bool {
	return true
}

func (c *ContextValueNode) GetCtn() ast_ifaces.Container {
	return c.Ctn
}

func (c *ContextValueNode) SetCtn(cc ast_ifaces.Container) {
	c.Ctn = cc
}

func (c *ContextValueNode) SetName(name string) {
	c.Node.Name = name
}

func (c *ContextValueNode) GetName() string {
	return c.Name
}
