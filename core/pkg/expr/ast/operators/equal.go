package operators

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/base"
)

type Equal struct {
	base.Container
	base.Node
}

func (e *Equal) Value() any {
	panic("not implemented")
}

func (e *Equal) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitEqual(eCtx, e)
}

func (e *Equal) TraceFullyRealized() bool {
	return false
}

func (e *Equal) Expr() string {
	return fmt.Sprintf("(%s == %s)", e.Params[0].Expr(), e.Params[1].Expr())
}

func (e *Equal) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(e)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s == %s)", e.Params[0].RealizedExpr(eCtx), e.Params[1].RealizedExpr(eCtx))
}

func (e *Equal) GetCtn() ast_ifaces.Container {
	return e.Ctn
}

func (e *Equal) SetCtn(c ast_ifaces.Container) {
	e.Ctn = c
}

func (e *Equal) GetLevel() (level int) {
	return e.Level
}

func (e *Equal) GetName() string {
	return e.Name
}
func (e *Equal) SetName(name string) {
	e.Name = name
}
