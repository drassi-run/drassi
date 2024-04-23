package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type ContextValueNode struct {
	NamedValue
	base.ExpressionNodeBs
}

func (c *ContextValueNode) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitContextValueNode(eCtx, c)
}

func (c *ContextValueNode) Value() any {
	panic("not implemented")
}

func (c *ContextValueNode) TraceFullyRealized() bool {
	return true
}

func (c *ContextValueNode) GetContainer() interfaces.IContainer {
	return c.Container
}

func (c *ContextValueNode) SetContainer(cc interfaces.IContainer) {
	c.Container = cc
}

func (c *ContextValueNode) SetName(name string) {
	c.ExpressionNodeBs.Name = name
}

func (c *ContextValueNode) GetName() string {
	return c.Name
}
