package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type ContainsFn struct {
	Fn
}

func (c *ContainsFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitContainsFn(eCtx, c)
}

func (c *ContainsFn) Value() any {
	panic("not implemented")
}

func (c *ContainsFn) TraceFullyRealized() bool {
	return false
}

//
// func (c *ContainsFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, c.Parameters()[0])
// 	if l.IsPrimitive() {
// 		lStr := l.ConvertToString()
// 		r := evaluator.EvaluateWithContext(eCtx, c.Parameters()[1])
// 		if r.IsPrimitive() {
// 			rStr := r.ConvertToString()
// 			return containsIgnoreCase(lStr, rStr)
// 		}
// 	}
// 	isCol, col := l.TryGetCollectionInterface()
// 	if isCol {
// 		if arr, isArr := col.(interfaces.IReadOnlyArray); isArr && arr.Count() > 0 {
// 			r := evaluator.EvaluateWithContext(eCtx, c.Parameters()[1])
// 			e := arr.Enumerator()
// 			for e.Next() {
// 				i := evaluator.CreateIntermediateResult(eCtx, e.Value())
// 				if r.AbstractEqual(i) {
// 					return true
// 				}
// 			}
// 		}
// 	}
// 	return false
// }

func containsIgnoreCase(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func (c *ContainsFn) ConvertToExpression() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *ContainsFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(c)
	if exist {
		return result
	}
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *ContainsFn) SetName(name string) {
	c.Name = name
}

func (c *ContainsFn) GetName() string {
	return c.Name
}

func (c *ContainsFn) GetContainer() expression.IContainer {
	return c.Container
}

func (c *ContainsFn) SetContainer(cc expression.IContainer) {
	c.Container = cc
}
