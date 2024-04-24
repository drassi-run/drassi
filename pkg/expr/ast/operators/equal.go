package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type Equal struct {
	base.Container
	base.Node
}

func (e *Equal) Value() any {
	panic("not implemented")
}

func (e *Equal) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitEqual(eCtx, e)
}

func (e *Equal) TraceFullyRealized() bool {
	return false
}

func (e *Equal) Expr() string {
	return fmt.Sprintf("(%s == %s)", e.Params[0].Expr(), e.Params[1].Expr())
}

func (e *Equal) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(e)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s == %s)", e.Params[0].RealizedExpr(eCtx), e.Params[1].RealizedExpr(eCtx))
}

func (e *Equal) GetCtn() interfaces.Container {
	return e.Ctn
}

func (e *Equal) SetCtn(c interfaces.Container) {
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
