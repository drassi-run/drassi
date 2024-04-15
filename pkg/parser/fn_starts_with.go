package parser

import (
	"fmt"
	"strings"
)

type StartsWithFn struct {
	Fn
}

func (s *StartsWithFn) traceFullyRealized() bool {
	return false
}

func (s *StartsWithFn) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, s.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluate(eCtx, s.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return startsWithIgnoreCase(lStr, rStr)
		}
	}
	return false
}

func startsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(suffix))
}

func (s *StartsWithFn) setName(name string) {
	s.name = name
}

func (s *StartsWithFn) getName() string {
	return s.name
}

func (s *StartsWithFn) getContainer() iContainer {
	return s.container
}

func (s *StartsWithFn) setContainer(c iContainer) {
	s.container = c
}

func (s *StartsWithFn) convertToExpression() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", s.getName(), strings.Join(params, ", "))
}

func (s *StartsWithFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(s)
	if exist {
		return result
	}
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", s.getName(), strings.Join(params, ", "))
}
