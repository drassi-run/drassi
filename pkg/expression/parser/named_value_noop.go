package parser

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

type noOpNamedValue struct {
	NamedValue
}

func (n *noOpNamedValue) Value() any {
	panic("not implemented")
}

func (n *noOpNamedValue) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitNoopNamedValue(eCtx, n)
}

func (n *noOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (n *noOpNamedValue) GetContainer() expression.IContainer {
	return n.Container
}

func (n *noOpNamedValue) SetContainer(cc expression.IContainer) {
	n.Container = cc
}

func (n *noOpNamedValue) SetName(name string) {
	n.ExpressionNodeBs.Name = name
}

func (n *noOpNamedValue) GetName() string {
	return n.Name
}
