package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Failure struct {
	Fn
}

func (f *Failure) Value() any {
	panic("not implemented")
}

func (f *Failure) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitFailureFn(eCtx, f)
}

func (f *Failure) SetName(name string) {
	f.Name = name
}

func (f *Failure) GetName() string {
	return f.Name
}

func (f *Failure) GetCtn() interfaces.Container {
	return f.Ctn
}

func (f *Failure) SetCtn(cc interfaces.Container) {
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

func (f *Failure) RealizedExpr(eCtx interfaces.Context) string {
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
