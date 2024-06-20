package ast

import (
	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/base"
	"drassi.run/core/pkg/expression/common"
)

type WildCard struct {
	base.Node
}

func (w *WildCard) Value() any {
	panic("not implemented")
}

func (w *WildCard) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitWildCard(eCtx, w)
}

func (w *WildCard) TraceFullyRealized() bool {
	return false
}

func (w *WildCard) Expr() string {
	return common.Wildcard
}

func (w *WildCard) RealizedExpr(eCtx ast_ifaces.Context) string {
	return common.Wildcard
}

func (w *WildCard) SetCtn(c ast_ifaces.Container) {
	w.Ctn = c
}

func (w *WildCard) GetCtn() ast_ifaces.Container {
	return w.Ctn
}
