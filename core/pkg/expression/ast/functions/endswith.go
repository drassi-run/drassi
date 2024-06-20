package functions

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
)

type EndsWith struct {
	Fn
}

func (e *EndsWith) Value() any {
	panic("not implemented")
}

func (e *EndsWith) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitEndsWithFn(eCtx, e)
}

func (e *EndsWith) TraceFullyRealized() bool {
	return false
}

func endsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(suffix))
}

func (e *EndsWith) Expr() string {
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", e.GetName(), strings.Join(params, ", "))
}

func (e *EndsWith) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(e)
	if exist {
		return result
	}
	params := make([]string, len(e.Parameters()))
	for i, param := range e.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", e.GetName(), strings.Join(params, ", "))
}

func (e *EndsWith) SetName(name string) {
	e.Name = name
}

func (e *EndsWith) GetName() string {
	return e.Name
}

func (e *EndsWith) GetCtn() ast_ifaces.Container {
	return e.Ctn
}

func (e *EndsWith) SetCtn(cc ast_ifaces.Container) {
	e.Ctn = cc
}
