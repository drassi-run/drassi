package parser

import (
	"fmt"
	"strings"
)

type AlwaysFn struct {
	Fn
	name string
}

func (a *AlwaysFn) evaluateCore(eCtx *EvaluationContext) any {
	return true
}

func (a *AlwaysFn) setName(name string) {
	a.name = name
}

func (a *AlwaysFn) getName() string {
	return a.name
}

func (a *AlwaysFn) getContainer() iContainer {
	return a.container
}

func (a *AlwaysFn) setContainer(c iContainer) {
	a.container = c
}

func (a *AlwaysFn) convertToExpression() string {
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", a.getName(), strings.Join(params, ", "))
}

func (a *AlwaysFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(a)
	if exist {
		return result
	}
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", a.getName(), strings.Join(params, ", "))
}

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expression is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expression expanded at run time. For example, consider
// the expression: eq(variables.publish, 'true'). The runtime-expanded expression may be: eq('true', 'true')
func (a *AlwaysFn) traceFullyRealized() bool {
	return true
}
