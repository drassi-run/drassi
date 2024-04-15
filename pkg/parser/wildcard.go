package parser

import (
	"github.com/dungdm93/drasi/pkg/parser/constants"
)

type WildCard struct {
	ExpressionNode
}

func (w *WildCard) traceFullyRealized() bool {
	return false
}

func (w *WildCard) convertToExpression() string {
	return constants.Wildcard
}

func (w *WildCard) convertToRealizedExpression(eCtx *EvaluationContext) string {
	return constants.Wildcard
}

func (w *WildCard) evaluateCore(eCtx *EvaluationContext) any {
	return constants.Wildcard
}
