package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
)

type Contains struct {
	Fn
}

func (c *Contains) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitContainsFn(eCtx, c)
}

func (c *Contains) Value() any {
	panic("not implemented")
}

func (c *Contains) TraceFullyRealized() bool {
	return false
}

func containsIgnoreCase(s string, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func (c *Contains) Expr() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *Contains) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(c)
	if exist {
		return result
	}
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.RealizedExpr(eCtx)
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *Contains) SetName(name string) {
	c.Name = name
}

func (c *Contains) GetName() string {
	return c.Name
}

func (c *Contains) GetCtn() ast_ifaces.Container {
	return c.Ctn
}

func (c *Contains) SetCtn(cc ast_ifaces.Container) {
	c.Ctn = cc
}
