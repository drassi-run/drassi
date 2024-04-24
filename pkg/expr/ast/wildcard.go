package ast

import (
	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type WildCard struct {
	base.Node
}

func (w *WildCard) Value() any {
	panic("not implemented")
}

func (w *WildCard) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitWildCard(eCtx, w)
}

func (w *WildCard) TraceFullyRealized() bool {
	return false
}

func (w *WildCard) Expr() string {
	return common.Wildcard
}

func (w *WildCard) RealizedExpr(eCtx interfaces.Context) string {
	return common.Wildcard
}

func (w *WildCard) SetCtn(c interfaces.Container) {
	w.Ctn = c
}

func (w *WildCard) GetCtn() interfaces.Container {
	return w.Ctn
}
