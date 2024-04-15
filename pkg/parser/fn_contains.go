package parser

import (
	"fmt"
	"strings"
)

type ContainsFn struct {
	Fn
}

func (c *ContainsFn) traceFullyRealized() bool {
	return false
}

func (c *ContainsFn) evaluateCore(eCtx *EvaluationContext) any {
	l := evaluate(eCtx, c.Parameters()[0])
	if l.IsPrimitive() {
		lStr := l.ConvertToString()
		r := evaluate(eCtx, c.Parameters()[1])
		if r.IsPrimitive() {
			rStr := r.ConvertToString()
			return containsIgnoreCase(lStr, rStr)
		}
	}
	isCol, col := l.TryGetCollectionInterface()
	if isCol {
		if arr, isArr := col.(IReadOnlyArray); isArr && arr.Count() > 0 {
			r := evaluate(eCtx, c.Parameters()[1])
			e := arr.Enumerator()
			for e.Next() {
				i := CreateIntermediateResult(eCtx, e.Value())
				if r.AbstractEqual(i) {
					return true
				}
			}
		}
	}
	return false
}

func containsIgnoreCase(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func (c *ContainsFn) convertToExpression() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", c.getName(), strings.Join(params, ", "))
}

func (c *ContainsFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(c)
	if exist {
		return result
	}
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", c.getName(), strings.Join(params, ", "))
}

func (c *ContainsFn) setName(name string) {
	c.name = name
}

func (c *ContainsFn) getName() string {
	return c.name
}

func (c *ContainsFn) getContainer() iContainer {
	return c.container
}

func (c *ContainsFn) setContainer(cc iContainer) {
	c.container = cc
}
