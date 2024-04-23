package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type EndsWithFn struct {
	Fn
}

func (e *EndsWithFn) Value() any {
	panic("not implemented")
}

func (e *EndsWithFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
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

func (e *EndsWithFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
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

func (e *EndsWithFn) GetContainer() interfaces.IContainer {
	return e.Container
}

func (e *EndsWithFn) SetContainer(cc interfaces.IContainer) {
	e.Container = cc
}
