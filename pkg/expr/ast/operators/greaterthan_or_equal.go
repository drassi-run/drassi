package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type GreaterThanOrEqual struct {
	base.Node
	base.Container
}

func (g *GreaterThanOrEqual) Value() any {
	panic("not implemented")
}

func (g *GreaterThanOrEqual) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitGreaterThanOrEqual(eCtx, g)
}

func (g *GreaterThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThanOrEqual) Expr() string {
	return fmt.Sprintf("(%s >= %s)", g.Params[0].Expr(), g.Params[1].Expr())
}

func (g *GreaterThanOrEqual) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s >= %s)", g.Params[0].RealizedExpr(eCtx), g.Params[1].RealizedExpr(eCtx))
}

func (g *GreaterThanOrEqual) GetCtn() interfaces.Container {
	return g.Ctn
}

func (g *GreaterThanOrEqual) SetCtn(c interfaces.Container) {
	g.Ctn = c
}

func (g *GreaterThanOrEqual) GetLevel() (level int) {
	return g.Level
}

func (g *GreaterThanOrEqual) GetName() string {
	return g.Name
}

func (g *GreaterThanOrEqual) SetName(name string) {
	g.Name = name
}
