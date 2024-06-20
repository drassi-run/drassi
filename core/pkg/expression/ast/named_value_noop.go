package ast

import (
	"drassi.run/core/pkg/expression/ast/ast_ifaces"
)

type NoOpNamedValue struct {
	NamedValue
}

func (n *NoOpNamedValue) Value() any {
	panic("not implemented")
}

func (n *NoOpNamedValue) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitNoopNamedValue(eCtx, n)
}

func (n *NoOpNamedValue) TraceFullyRealized() bool {
	return true
}

func (n *NoOpNamedValue) GetCtn() ast_ifaces.Container {
	return n.Ctn
}

func (n *NoOpNamedValue) SetCtn(cc ast_ifaces.Container) {
	n.Ctn = cc
}

func (n *NoOpNamedValue) SetName(name string) {
	n.Node.Name = name
}

func (n *NoOpNamedValue) GetName() string {
	return n.Name
}
