package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
	"github.com/dungdm93/drasi/pkg/expression/shared"
)

type WildCard struct {
	base.ExpressionNodeBs
}

func (w *WildCard) Value() any {
	panic("not implemented")
}

func (w *WildCard) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitWildCard(eCtx, w)
}

func (w *WildCard) TraceFullyRealized() bool {
	return false
}

func (w *WildCard) ConvertToExpression() string {
	return shared.Wildcard
}

func (w *WildCard) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	return shared.Wildcard
}
