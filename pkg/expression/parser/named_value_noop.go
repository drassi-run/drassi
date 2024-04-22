package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type NoOpNamedValue struct {
	NamedValue
}

func (a *NoOpNamedValue) Value() any {
	panic("not implemented")
}

func (a *NoOpNamedValue) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitNoopNamedValue(eCtx, a)
}

// func (c *NoOpNamedValue) EvaluateCore(eCtx interfaces.IEvaluationContext) (result any) {
//
// 	return nil
// }

func (c *NoOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (c *NoOpNamedValue) GetContainer() interfaces.IContainer {
	return c.Container
}

func (c *NoOpNamedValue) SetContainer(cc interfaces.IContainer) {
	c.Container = cc
}

func (c *NoOpNamedValue) SetName(name string) {
	c.ExpressionNodeBs.Name = name
}

func (c *NoOpNamedValue) GetName() string {
	return c.Name
}
