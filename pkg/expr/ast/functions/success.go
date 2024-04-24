package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Success struct {
	Fn
}

func (s *Success) Value() any {
	panic("not implemented")
}

func (s *Success) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitSuccessFn(eCtx, s)
}
func (s *Success) SetName(name string) {
	s.Name = name
}

func (s *Success) GetName() string {
	return s.Name
}

func (s *Success) GetCtn() interfaces.Container {
	return s.Ctn
}

func (s *Success) SetCtn(cc interfaces.Container) {
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

func (s *Success) RealizedExpr(eCtx interfaces.Context) string {
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
