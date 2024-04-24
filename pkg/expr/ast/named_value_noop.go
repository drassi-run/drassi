package ast

import (
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type noOpNamedValue struct {
	NamedValue
}

func (n *noOpNamedValue) Value() any {
	panic("not implemented")
}

func (n *noOpNamedValue) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitNoopNamedValue(eCtx, n)
}

func (n *noOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (n *noOpNamedValue) GetCtn() interfaces.Container {
	return n.Ctn
}

func (n *noOpNamedValue) SetCtn(cc interfaces.Container) {
	n.Ctn = cc
}

func (n *noOpNamedValue) SetName(name string) {
	n.Node.Name = name
}

func (n *noOpNamedValue) GetName() string {
	return n.Name
}
