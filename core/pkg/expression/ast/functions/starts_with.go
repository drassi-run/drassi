package functions

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
)

type StartsWith struct {
	Fn
}

func (s *StartsWith) Value() any {
	panic("not implemented")
}

func (s *StartsWith) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitStartsWithFn(eCtx, s)
}

func (s *StartsWith) TraceFullyRealized() bool {
	return false
}

func startsWithIgnoreCase(str string, suffix string) bool {
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(suffix))
}

func (s *StartsWith) SetName(name string) {
	s.Name = name
}

func (s *StartsWith) GetName() string {
	return s.Name
}

func (s *StartsWith) GetCtn() ast_ifaces.Container {
	return s.Ctn
}

func (s *StartsWith) SetCtn(c ast_ifaces.Container) {
	s.Ctn = c
}

func (s *StartsWith) Expr() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}

func (s *StartsWith) RealizedExpr(eCtx ast_ifaces.Context) string {
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
