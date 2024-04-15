package parser

import (
	"fmt"
	"strings"
)

type And struct {
	Container
	ExpressionNode
}

func (a *And) convertToExpression() string {
	expressions := make([]string, len(a.params))
	for i, param := range a.params {
		expressions[i] = param.convertToExpression()
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(a)
	if exist {
		return result
	}
	expressions := make([]string, len(a.params))
	for i, param := range a.params {
		expressions[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) evaluateCore(eCtx *EvaluationContext) any {
	result := &EvaluationResult{}
	for _, param := range a.params {
		result = evaluate(eCtx, param)
		if result.IsFalsy() {
			return result.Value()
		}
	}
	return result.Value()
}

func (a *And) traceFullyRealized() bool {
	return false
}

func (a *And) getContainer() iContainer {
	return a.container
}

func (a *And) setContainer(c iContainer) {
	a.container = c
}

func (a *And) getLevel() (level int) {
	return a.level
}

func (a *And) getName() string {
	return a.name
}
func (a *And) setName(name string) {
	a.name = name
}
