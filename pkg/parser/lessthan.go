package parser

import (
	"fmt"
)

type LessThan struct {
	ExpressionNode
	Container
}

func (l *LessThan) traceFullyRealized() bool {
	return false
}

func (l *LessThan) convertToExpression() string {
	return fmt.Sprintf("(%s < %s)", l.params[0].convertToExpression(), l.params[1].convertToExpression())
}

func (l *LessThan) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s < %s)", l.params[0].convertToRealizedExpression(eCtx), l.params[1].convertToRealizedExpression(eCtx))
}

func (l *LessThan) evaluateCore(eCtx *EvaluationContext) any {
	left := evaluate(eCtx, l.params[0])
	right := evaluate(eCtx, l.params[1])
	return left.AbstractLessThan(right)
}

func (l *LessThan) getContainer() iContainer {
	return l.container
}

func (l *LessThan) setContainer(c iContainer) {
	l.container = c
}

func (l *LessThan) getLevel() (level int) {
	return l.level
}

func (l *LessThan) getName() string {
	return l.name
}
func (l *LessThan) setName(name string) {
	l.name = name
}
