package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
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
func (f *JoinFn) TraceFullyRealized() bool {
	return true
}

func (a *JoinFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitJoinFn(eCtx, a)
}

func (f *JoinFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	items := evaluator.EvaluateWithContext(eCtx, f.Parameters()[0])
	isCol, col := items.TryGetCollectionInterface()
	if isCol {
		arr, isArr := col.(interfaces.IReadOnlyArray)
		if isArr && arr.Count() > 0 {
			var result strings.Builder
			item := arr.GetValue(0)
			itemResult := evaluator.CreateIntermediateResult(eCtx, item)
			itemStr := itemResult.ConvertToString()
			_, err := result.WriteString(itemStr)
			if err != nil {
				return ""
			}
			if arr.Count() > 1 {
				separator := ","
				if len(f.Parameters()) > 1 {
					separatorResult := evaluator.EvaluateWithContext(eCtx, f.Parameters()[1])
					if separatorResult.IsPrimitive() {
						separator = separatorResult.ConvertToString()
					}
				}
				e := arr.Enumerator()
				for e.Next() {
					result.WriteString(separator)
					nextItem := e.Value()
					nextItemResult := evaluator.CreateIntermediateResult(eCtx, nextItem)
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

func (f *JoinFn) SetName(name string) {
	f.Name = name
}

func (f *JoinFn) GetName() string {
	return f.Name
}

func (f *JoinFn) GetContainer() interfaces.IContainer {
	return f.Container
}

func (f *JoinFn) SetContainer(c interfaces.IContainer) {
	f.Container = c
}

func (f *JoinFn) ConvertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *JoinFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}
