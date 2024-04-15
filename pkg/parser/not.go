package parser

import (
	"fmt"
)

type Not struct {
	ExpressionNode
	Container
}

func (n *Not) traceFullyRealized() bool {
	return false
}

func (n *Not) convertToExpression() string {
	return fmt.Sprintf("!%s", n.params[0].convertToExpression())
}

func (n *Not) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("!%s", n.params[0].convertToRealizedExpression(eCtx))
}

func (n *Not) evaluateCore(eCtx *EvaluationContext) any {
	return evaluate(eCtx, n.params[0]).IsFalsy()
}

func (n *Not) getContainer() iContainer {
	return n.container
}

func (n *Not) setContainer(c iContainer) {
	n.container = c
}

func (n *Not) getLevel() (level int) {
	return n.level
}

func (n *Not) getName() string {
	return n.name
}
func (n *Not) setName(name string) {
	n.name = name
}
