package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type AlwaysFn struct {
	Fn
	name string
}

func (a *AlwaysFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitAlwaysFn(eCtx, a)
}

func (a *AlwaysFn) Value() any {
	panic("not implemented")
}

// func (a *AlwaysFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	return true
// }

func (a *AlwaysFn) SetName(name string) {
	a.name = name
}

func (a *AlwaysFn) GetName() string {
	return a.name
}

func (a *AlwaysFn) GetContainer() interfaces.IContainer {
	return a.Container
}

func (a *AlwaysFn) SetContainer(c interfaces.IContainer) {
	a.Container = c
}

func (a *AlwaysFn) ConvertToExpression() string {
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(params, ", "))
}

func (a *AlwaysFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(a)
	if exist {
		return result
	}
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(params, ", "))
}

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expression is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expression expanded at run time. For example, consider
// the expression: eq(variables.publish, 'true'). The runtime-expanded expression may be: eq('true', 'true')
func (a *AlwaysFn) TraceFullyRealized() bool {
	return true
}
