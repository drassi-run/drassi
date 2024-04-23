package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type EndsWithFn struct {
	Fn
}

func (e *EndsWithFn) Value() any {
	panic("not implemented")
}

func (e *EndsWithFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitEndsWithFn(eCtx, e)
}

func (e *EndsWithFn) TraceFullyRealized() bool {
	return false
}

//
// func (e *EndsWithFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, e.Parameters()[0])
// 	if l.IsPrimitive() {
// 		lStr := l.ConvertToString()
// 		r := evaluator.EvaluateWithContext(eCtx, e.Parameters()[1])
// 		if r.IsPrimitive() {
// 			rStr := r.ConvertToString()
// 			return endsWithIgnoreCase(lStr, rStr)
// 		}
// 	}
// 	return false
// }

func endsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

func (e *EndsWithFn) ConvertToExpression() string {
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", e.GetName(), strings.Join(params, ", "))
}

func (e *EndsWithFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(e)
	if exist {
		return result
	}
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", e.GetName(), strings.Join(params, ", "))
}

func (e *EndsWithFn) SetName(name string) {
	e.Name = name
}

func (e *EndsWithFn) GetName() string {
	return e.Name
}

func (e *EndsWithFn) GetContainer() expression.IContainer {
	return e.Container
}

func (e *EndsWithFn) SetContainer(cc expression.IContainer) {
	e.Container = cc
}
