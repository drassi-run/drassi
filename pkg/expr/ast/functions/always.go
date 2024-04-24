package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Always struct {
	Fn
	name string
}

func (a *Always) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitAlwaysFn(eCtx, a)
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

func (a *Always) GetCtn() interfaces.Container {
	return a.Ctn
}

func (a *Always) SetCtn(c interfaces.Container) {
	a.Ctn = c
}

func (a *Always) Expr() string {
	params := make([]string, len(a.Parameters()))
	for i, param := range a.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", a.GetName(), strings.Join(params, ", "))
}

func (a *Always) RealizedExpr(eCtx interfaces.Context) string {
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

// Generally this should not be overridden. True indicates the result of the node is traced as part of the "expanded"
// (i.e. "realized") trace information. Otherwise, the node expr is printed, and parameters to the node may or
// may not be fully realized - depending on each respective parameter's trace-fully-realized setting.
//
// The purpose is so the end user can understand how their expr expanded at run time. For example, consider
// the expr: eq(variables.publish, 'true'). The runtime-expanded expr may be: eq('true', 'true')
func (a *Always) TraceFullyRealized() bool {
	return true
}
