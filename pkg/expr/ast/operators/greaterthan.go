package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/base"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

type GreaterThan struct {
	base.Node
	base.Container
}

func (g *GreaterThan) Value() any {
	panic("not implemented")
}

func (g *GreaterThan) Accept(eCtx interfaces.Context, v interfaces.Visitor) any {
	return v.VisitGreaterThan(eCtx, g)
}

func (g *GreaterThan) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThan) Expr() string {
	return fmt.Sprintf("(%s > %s)", g.Params[0].Expr(), g.Params[1].Expr())
}

func (g *GreaterThan) RealizedExpr(eCtx interfaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s > %s)", g.Params[0].RealizedExpr(eCtx), g.Params[1].RealizedExpr(eCtx))
}

func (g *GreaterThan) GetCtn() interfaces.Container {
	return g.Ctn
}

func (g *GreaterThan) SetCtn(c interfaces.Container) {
	g.Ctn = c
}

func (g *GreaterThan) GetLevel() (level int) {
	return g.Level
}

func (g *GreaterThan) GetName() string {
	return g.Name
}
func (g *GreaterThan) SetName(name string) {
	g.Name = name
}
