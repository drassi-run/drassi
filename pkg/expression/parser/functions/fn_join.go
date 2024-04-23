package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type JoinFn struct {
	Fn
}

func (j *JoinFn) Value() any {
	panic("not implemented")
}

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expression is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expression expanded at run time. For example, consider
// the expression: eq(variables.publish, 'true'). The runtime-expanded expression may be: eq('true', 'true')
func (j *JoinFn) TraceFullyRealized() bool {
	return true
}

func (j *JoinFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitJoinFn(eCtx, j)
}

//
// func (f *JoinFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	items := evaluator.EvaluateWithContext(eCtx, f.Parameters()[0])
// 	isCol, col := items.IsCollection()
// 	if isCol {
// 		arr, isArr := col.(interfaces.ReadOnlyArray)
// 		if isArr && arr.Count() > 0 {
// 			var result strings.Builder
// 			item := arr.GetValue(0)
// 			itemResult := evaluator.CreateIntermediateResult(eCtx, item)
// 			itemStr := itemResult.ConvertToString()
// 			_, err := result.WriteString(itemStr)
// 			if err != nil {
// 				return ""
// 			}
// 			if arr.Count() > 1 {
// 				separator := ","
// 				if len(f.Parameters()) > 1 {
// 					separatorResult := evaluator.EvaluateWithContext(eCtx, f.Parameters()[1])
// 					if separatorResult.IsPrimitive() {
// 						separator = separatorResult.ConvertToString()
// 					}
// 				}
// 				e := arr.Enumerator()
// 				for e.Next() {
// 					result.WriteString(separator)
// 					nextItem := e.Value()
// 					nextItemResult := evaluator.CreateIntermediateResult(eCtx, nextItem)
// 					nextItemStr := nextItemResult.ConvertToString()
// 					result.WriteString(nextItemStr)
// 				}
// 			}
// 			return result.ValueKindString()
// 		}
// 	}
// 	if items.IsPrimitive() {
// 		return items.ConvertToString()
// 	}
// 	return ""
// }

func (j *JoinFn) SetName(name string) {
	j.Name = name
}

func (j *JoinFn) GetName() string {
	return j.Name
}

func (j *JoinFn) GetContainer() expression.IContainer {
	return j.Container
}

func (j *JoinFn) SetContainer(c expression.IContainer) {
	j.Container = c
}

func (j *JoinFn) ConvertToExpression() string {
	params := make([]string, len(j.Parameters()))
	for i, param := range j.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", j.GetName(), strings.Join(params, ", "))
}

func (j *JoinFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(j)
	if exist {
		return result
	}
	params := make([]string, len(j.Parameters()))
	for i, param := range j.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", j.GetName(), strings.Join(params, ", "))
}
