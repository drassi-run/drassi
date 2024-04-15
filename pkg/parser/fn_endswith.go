package parser

import (
	"fmt"
	"strings"
)

type EndsWithFn struct {
	Fn
}

func (e *EndsWithFn) traceFullyRealized() bool {
	return false
}

func (e *EndsWithFn) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, e.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluate(eCtx, e.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return endsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func endsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

func (e *EndsWithFn) convertToExpression() string {
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", e.getName(), strings.Join(params, ", "))
}

func (e *EndsWithFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(e)
	if exist {
		return result
	}
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", e.getName(), strings.Join(params, ", "))
}

func (e *EndsWithFn) setName(name string) {
	e.name = name
}

func (e *EndsWithFn) getName() string {
	return e.name
}

func (e *EndsWithFn) getContainer() iContainer {
	return e.container
}
