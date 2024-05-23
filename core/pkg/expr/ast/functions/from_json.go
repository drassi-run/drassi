package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
)

type FromJson struct {
	Fn
}

func (f *FromJson) Value() any {
	panic("not implemented")
}

func (f *FromJson) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitFromJsonFn(eCtx, f)
}

func (f *FromJson) TraceFullyRealized() bool {
	return false
}

func (f *FromJson) SetName(name string) {
	f.Name = name
}

func (f *FromJson) GetName() string {
	return f.Name
}

func (f *FromJson) GetCtn() ast_ifaces.Container {
	return f.Ctn
}

func (f *FromJson) SetCtn(c ast_ifaces.Container) {
	f.Ctn = c
}

func (f *FromJson) Expr() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *FromJson) RealizedExpr(eCtx ast_ifaces.Context) string {
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
