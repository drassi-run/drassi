package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Cancelled struct {
	Fn
}

func (c *Cancelled) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
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

func (c *Cancelled) RealizedExpr(eCtx interfaces.Context) string {
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

func (c *Cancelled) GetCtn() interfaces.Container {
	return c.Ctn
}

func (c *Cancelled) SetCtn(cc interfaces.Container) {
	c.Ctn = cc
}

func (c *Cancelled) TraceFullyRealized() bool {
	return false
}
