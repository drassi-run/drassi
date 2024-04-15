package parser

import (
	"fmt"
)

type GreaterThan struct {
	ExpressionNode
	Container
}

func (g *GreaterThan) traceFullyRealized() bool {
	return false
}

func (g *GreaterThan) convertToExpression() string {
	return fmt.Sprintf("(%s > %s)", g.params[0].convertToExpression(), g.params[1].convertToExpression())
}

func (g *GreaterThan) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s > %s)", g.params[0].convertToRealizedExpression(eCtx), g.params[1].convertToRealizedExpression(eCtx))
}

func (g *GreaterThan) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, g.params[0])
	r := evaluate(eCtx, g.params[1])
	return l.AbstractGreaterThan(r)
}

func (g *GreaterThan) getContainer() iContainer {
	return g.container
}

func (g *GreaterThan) setContainer(c iContainer) {
	g.container = c
}

func (g *GreaterThan) getLevel() (level int) {
	return g.level
}

func (g *GreaterThan) getName() string {
	return g.name
}
func (g *GreaterThan) setName(name string) {
	g.name = name
}
