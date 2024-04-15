package parser

import (
	"fmt"
)

type NotEqual struct {
	ExpressionNode
	Container
}

func (n *NotEqual) traceFullyRealized() bool {
	return false
}

func (n *NotEqual) convertToExpression() string {
	return fmt.Sprintf("%s != %s", n.params[0].convertToExpression(), n.params[1].convertToExpression())
}

func (n *NotEqual) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("%s != %s", n.params[0].convertToRealizedExpression(eCtx), n.params[1].convertToRealizedExpression(eCtx))
}

func (n *NotEqual) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, n.params[0])
	r := evaluate(eCtx, n.params[1])
	return l.AbstractNotEqual(r)
}

func (n *NotEqual) getContainer() iContainer {
	return n.container
}

func (n *NotEqual) setContainer(c iContainer) {
	n.container = c
}

func (n *NotEqual) getLevel() (level int) {
	return n.level
}

func (n *NotEqual) getName() string {
	return n.name
}
func (n *NotEqual) setName(name string) {
	n.name = name
}
