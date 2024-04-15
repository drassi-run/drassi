package parser

import (
	"fmt"
	"strings"
)

type NoOpFn struct {
	ExpressionNode
	Fn
}

func (n *NoOpFn) evaluateCore(eCtx *EvaluationContext) any {
	return nil
}

func (n *NoOpFn) traceFullyRealized() bool {
	return false
}

func (n *NoOpFn) getContainer() iContainer {
	return n.container
}

func (n *NoOpFn) setContainer(c iContainer) {
	n.container = c
}

func (n *NoOpFn) getLevel() (level int) {
	return n.level
}

func (n *NoOpFn) getName() string {
	return n.name
}
func (n *NoOpFn) setName(name string) {
	n.name = name
}

func (n *NoOpFn) convertToExpression() string {
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", n.getName(), strings.Join(params, ", "))
}

func (n *NoOpFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(n)
	if exist {
		return result
	}
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", n.getName(), strings.Join(params, ", "))
}
