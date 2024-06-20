package functions

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
)

type Failure struct {
	Fn
}

func (f *Failure) Value() any {
	panic("not implemented")
}

func (f *Failure) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitFailureFn(eCtx, f)
}

func (f *Failure) SetName(name string) {
	f.Name = name
}

func (f *Failure) GetName() string {
	return f.Name
}

func (f *Failure) GetCtn() ast_ifaces.Container {
	return f.Ctn
}

func (f *Failure) SetCtn(cc ast_ifaces.Container) {
	f.Ctn = cc
}

func (f *Failure) TraceFullyRealized() bool {
	return false
}

func (f *Failure) Expr() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *Failure) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}
