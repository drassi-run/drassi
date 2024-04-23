package parser

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

type NoOpNamedValue struct {
	NamedValue
}

func (n *NoOpNamedValue) Value() any {
	panic("not implemented")
}

func (n *NoOpNamedValue) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitNoopNamedValue(eCtx, n)
}

func (n *NoOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (n *NoOpNamedValue) GetContainer() expression.IContainer {
	return n.Container
}

func (n *NoOpNamedValue) SetContainer(cc expression.IContainer) {
	n.Container = cc
}

func (n *NoOpNamedValue) SetName(name string) {
	n.ExpressionNodeBs.Name = name
}

func (n *NoOpNamedValue) GetName() string {
	return n.Name
}
