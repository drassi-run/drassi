package parser

import (
	"fmt"
)

type LessThanOrEqual struct {
	ExpressionNode
	Container
}

func (l *LessThanOrEqual) traceFullyRealized() bool {
	return false
}

func (l *LessThanOrEqual) convertToExpression() string {
	return fmt.Sprintf("(%s <= %s)", l.params[0].convertToExpression(), l.params[1].convertToExpression())
}

func (l *LessThanOrEqual) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s <= %s)", l.params[0].convertToRealizedExpression(eCtx), l.params[1].convertToRealizedExpression(eCtx))
}

func (l *LessThanOrEqual) evaluateCore(eCtx *EvaluationContext) any {
	left := evaluate(eCtx, l.params[0])
	right := evaluate(eCtx, l.params[1])
	return left.AbstractLessThanOrEqual(right)
}

func (l *LessThanOrEqual) getContainer() iContainer {
	return l.container
}

func (l *LessThanOrEqual) setContainer(c iContainer) {
	l.container = c
}

func (l *LessThanOrEqual) getLevel() (level int) {
	return l.level
}

func (l *LessThanOrEqual) getName() string {
	return l.name
}
func (l *LessThanOrEqual) setName(name string) {
	l.name = name
}
