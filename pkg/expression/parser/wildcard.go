package parser

import (
	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type WildCard struct {
	base.ExpressionNodeBs
}

func (w *WildCard) Value() any {
	panic("not implemented")
}

func (w *WildCard) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitWildCard(eCtx, w)
}

func (w *WildCard) TraceFullyRealized() bool {
	return false
}

func (w *WildCard) ConvertToExpression() string {
	return expression.Wildcard
}

func (w *WildCard) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	return expression.Wildcard
}

func (w *WildCard) SetContainer(c expression.IContainer) {
	w.Container = c
}

func (w *WildCard) GetContainer() expression.IContainer {
	return w.Container
}
