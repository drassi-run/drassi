package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type Success struct {
	Fn
}

func (s *Success) Value() any {
	panic("not implemented")
}

func (s *Success) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitSuccessFn(eCtx, s)
}
func (s *Success) SetName(name string) {
	s.Name = name
}

func (s *Success) GetName() string {
	return s.Name
}

func (s *Success) GetCtn() ast_ifaces.Container {
	return s.Ctn
}

func (s *Success) SetCtn(cc ast_ifaces.Container) {
	s.Ctn = cc
}

func (s *Success) TraceFullyRealized() bool {
	return false
}

func (s *Success) Expr() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}

func (s *Success) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(s)
	if exist {
		return result
	}
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}
