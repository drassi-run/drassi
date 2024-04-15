package parser

import (
	"fmt"
)

type GreaterThanOrEqual struct {
	ExpressionNode
	Container
}

func (g *GreaterThanOrEqual) traceFullyRealized() bool {
	return false
}

func (g *GreaterThanOrEqual) convertToExpression() string {
	return fmt.Sprintf("(%s >= %s)", g.params[0].convertToExpression(), g.params[1].convertToExpression())
}

func (g *GreaterThanOrEqual) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s >= %s)", g.params[0].convertToRealizedExpression(eCtx), g.params[1].convertToRealizedExpression(eCtx))
}

func (g *GreaterThanOrEqual) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, g.params[0])
	r := evaluate(eCtx, g.params[1])
	return l.AbstractGreaterThan(r)
}

func (g *GreaterThanOrEqual) getContainer() iContainer {
	return g.container
}

func (g *GreaterThanOrEqual) setContainer(c iContainer) {
	g.container = c
}

func (g *GreaterThanOrEqual) getLevel() (level int) {
	return g.level
}

func (g *GreaterThanOrEqual) getName() string {
	return g.name
}
func (g *GreaterThanOrEqual) setName(name string) {
	g.name = name
}
