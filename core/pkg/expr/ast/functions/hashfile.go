package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
)

type HashFile struct {
	Fn
}

func (f *HashFile) Value() any {
	panic("not implemented")
}

func (f *HashFile) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitHashfileFn(eCtx, f)
}

func (f *HashFile) TraceFullyRealized() bool {
	return false
}

func (f *HashFile) SetName(name string) {
	f.Name = name
}

func (f *HashFile) GetName() string {
	return f.Name
}

func (f *HashFile) GetCtn() ast_ifaces.Container {
	return f.Ctn
}

func (f *HashFile) SetCtn(c ast_ifaces.Container) {
	f.Ctn = c
}

func (f *HashFile) Expr() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *HashFile) RealizedExpr(eCtx ast_ifaces.Context) string {
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
