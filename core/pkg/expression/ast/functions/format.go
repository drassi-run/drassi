package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type Format struct {
	Fn
}

func (f *Format) Value() any {
	panic("not implemented")
}

func (f *Format) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitFormatFn(eCtx, f)
}

func (f *Format) TraceFullyRealized() bool {
	return false
}

func (f *Format) SetName(name string) {
	f.Name = name
}

func (f *Format) GetName() string {
	return f.Name
}

func (f *Format) GetCtn() ast_ifaces.Container {
	return f.Ctn
}

func (f *Format) SetCtn(c ast_ifaces.Container) {
	f.Ctn = c
}

func (f *Format) Expr() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *Format) RealizedExpr(eCtx ast_ifaces.Context) string {
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
