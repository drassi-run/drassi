package functions

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
)

type Always struct {
	Fn
	name string
}

func (a *Always) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitAlwaysFn(a)
}

func (a *Always) Value() any {
	panic("not implemented")
}

func (a *Always) SetName(name string) {
	a.name = name
}

func (a *Always) GetName() string {
	return a.name
}

func (a *Always) GetCtn() ast_ifaces.Container {
	return a.Ctn
}

func (a *Always) SetCtn(c ast_ifaces.Container) {
	a.Ctn = c
}

func (a *Always) Expr() string {
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(params, ", "))
}

func (a *Always) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(a)
	if exist {
		return result
	}
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(params, ", "))
}

// True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expr is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expr expanded at run time. For example, consider
// the expr: eq(variables.publish, 'true'). The runtime-expanded expr may be: eq('true', 'true')
func (a *Always) TraceFullyRealized() bool {
	return true
}
