package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type NoOpNamedValue struct {
	NamedValue
}

func (n *NoOpNamedValue) Value() any {
	panic("not implemented")
}

func (n *NoOpNamedValue) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitNoopNamedValue(eCtx, n)
}

func (n *NoOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (n *NoOpNamedValue) GetContainer() interfaces.IContainer {
	return n.Container
}

func (n *NoOpNamedValue) SetContainer(cc interfaces.IContainer) {
	n.Container = cc
}

func (n *NoOpNamedValue) SetName(name string) {
	n.ExpressionNodeBs.Name = name
}

func (n *NoOpNamedValue) GetName() string {
	return n.Name
}
