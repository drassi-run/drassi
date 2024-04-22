package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type WildCard struct {
	base.ExpressionNodeBase
}

func (a *WildCard) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitWildCard(eCtx, a)
}

func (w *WildCard) TraceFullyRealized() bool {
	return false
}

func (w *WildCard) ConvertToExpression() string {
	return constants.Wildcard
}

func (w *WildCard) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	return constants.Wildcard
}

func (w *WildCard) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	return constants.Wildcard
}
