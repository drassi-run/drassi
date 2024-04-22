package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type StartsWithFn struct {
	Fn
}

func (s *StartsWithFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitStartsWithFn(eCtx, s)
}

func (s *StartsWithFn) TraceFullyRealized() bool {
	return false
}

func (s *StartsWithFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	l := evaluator.EvaluateWithContext(eCtx, s.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluator.EvaluateWithContext(eCtx, s.Parameters()[1])
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

func (s *StartsWithFn) SetName(name string) {
	s.Name = name
}

func (s *StartsWithFn) GetName() string {
	return s.Name
}

func (s *StartsWithFn) GetContainer() interfaces.IContainer {
	return s.Container
}

func (s *StartsWithFn) SetContainer(c interfaces.IContainer) {
	s.Container = c
}

func (s *StartsWithFn) ConvertToExpression() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}

func (s *StartsWithFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(s)
	if exist {
		return result
	}
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}
