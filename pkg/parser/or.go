package parser

import (
	"fmt"
)

type Or struct {
	ExpressionNode
	Container
}

func (o *Or) traceFullyRealized() bool {
	return false
}

func (o *Or) convertToExpression() string {
	return fmt.Sprintf("%s || %s", o.params[0].convertToExpression(), o.params[1].convertToExpression())
}

func (o *Or) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.params[0].convertToRealizedExpression(eCtx), o.params[1].convertToRealizedExpression(eCtx))
}

func (o *Or) evaluateCore(eCtx *EvaluationContext) any {
	var result *EvaluationResult
	for _, param := range o.params {
		result = evaluate(eCtx, param)
		if result.IsTruthy() {
			break
		}
	}
	if result == nil {
		return nil
	}
	return result.value
}

func (o *Or) getContainer() iContainer {
	return o.container
}

func (o *Or) setContainer(c iContainer) {
	o.container = c
}

func (o *Or) getLevel() (level int) {
	return o.level
}

func (o *Or) getName() string {
	return o.name
}
func (o *Or) setName(name string) {
	o.name = name
}
