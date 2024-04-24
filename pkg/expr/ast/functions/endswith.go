package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type EndsWith struct {
	Fn
}

func (e *EndsWith) Value() any {
	panic("not implemented")
}

func (e *EndsWith) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
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

func (e *EndsWith) RealizedExpr(eCtx interfaces.Context) string {
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

func (e *EndsWith) GetCtn() interfaces.Container {
	return e.Ctn
}

func (e *EndsWith) SetCtn(cc interfaces.Container) {
	e.Ctn = cc
}
