package parser

import (
	"fmt"
	"strings"
)

type JoinFn struct {
	Fn
}

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expression is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expression expanded at run time. For example, consider
// the expression: eq(variables.publish, 'true'). The runtime-expanded expression may be: eq('true', 'true')
func (f *JoinFn) traceFullyRealized() bool {
	return true
}

func (f *JoinFn) evaluateCore(eCtx *EvaluationContext) any {
	items := evaluate(eCtx, f.Parameters()[0])
	isCol, col := items.TryGetCollectionInterface()
	if isCol {
		arr, isArr := col.(IReadOnlyArray)
		if isArr && arr.Count() > 0 {
			var result strings.Builder
			item := arr.GetValue(0)
			itemResult := CreateIntermediateResult(eCtx, item)
			itemStr := itemResult.ConvertToString()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if arr.Count() > 1 {
				separator := ","
				if len(f.Parameters()) > 1 {
					separatorResult := evaluate(eCtx, f.Parameters()[1])
					if separatorResult.IsPrimitive() {
						separator = separatorResult.ConvertToString()
					}
				}
				e := arr.Enumerator()
				for e.Next() {
					result.WriteString(separator)
					nextItem := e.Value()
					nextItemResult := CreateIntermediateResult(eCtx, nextItem)
					nextItemStr := nextItemResult.ConvertToString()
					result.WriteString(nextItemStr)
				}
			}
			return result.String()
		}
	}
	if items.IsPrimitive() {
		return items.ConvertToString()
	}
	return ""
}

func (f *JoinFn) setName(name string) {
	f.name = name
}

func (f *JoinFn) getName() string {
	return f.name
}

func (f *JoinFn) getContainer() iContainer {
	return f.container
}

func (f *JoinFn) setContainer(c iContainer) {
	f.container = c
}

func (f *JoinFn) convertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}

func (f *JoinFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}
