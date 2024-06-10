package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
)

type Cancelled struct {
	Fn
}

func (c *Cancelled) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitCancelledFn(eCtx, c)
}

func (c *Cancelled) Value() any {
	panic("not implemented")
}

func (c *Cancelled) Expr() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.Expr()
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *Cancelled) RealizedExpr(eCtx ast_ifaces.Context) string {
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

func (c *Cancelled) SetName(name string) {
	c.Name = name
}

func (c *Cancelled) GetName() string {
	return c.Name
}

func (c *Cancelled) GetCtn() ast_ifaces.Container {
	return c.Ctn
}

func (c *Cancelled) SetCtn(cc ast_ifaces.Container) {
	c.Ctn = cc
}

func (c *Cancelled) TraceFullyRealized() bool {
	return false
}
