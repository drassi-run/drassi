package parser

import (
	"fmt"
)

type Equal struct {
	Container
	ExpressionNode
}

func (e *Equal) traceFullyRealized() bool {
	return false
}

func (e *Equal) convertToExpression() string {
	return fmt.Sprintf("(%s == %s)", e.params[0].convertToExpression(), e.params[1].convertToExpression())
}

func (e *Equal) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(e)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s == %s)", e.params[0].convertToRealizedExpression(eCtx), e.params[1].convertToRealizedExpression(eCtx))
}

func (e *Equal) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, e.params[0])
	r := evaluate(eCtx, e.params[1])
	return l.AbstractEqual(r)
}

func (e *Equal) getContainer() iContainer {
	return e.container
}

func (e *Equal) setContainer(c iContainer) {
	e.container = c
}

func (e *Equal) getLevel() (level int) {
	return e.level
}

func (e *Equal) getName() string {
	return e.name
}
func (e *Equal) setName(name string) {
	e.name = name
}
