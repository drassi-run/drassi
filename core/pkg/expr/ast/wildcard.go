package ast

import (
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/base"
	"github.com/dungdm93/drassi/core/pkg/expr/common"
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
