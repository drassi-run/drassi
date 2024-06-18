package operators

import (
	"fmt"

	"github.com/dungdm93/drassi/core/pkg/expression/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expression/ast/base"
)

type GreaterThanOrEqual struct {
	base.Node
	base.Container
}

func (g *GreaterThanOrEqual) Value() any {
	panic("not implemented")
}

func (g *GreaterThanOrEqual) Accept(eCtx ast_ifaces.Context, v ast_ifaces.Visitor) any {
	return v.VisitGreaterThanOrEqual(eCtx, g)
}

func (g *GreaterThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThanOrEqual) Expr() string {
	return fmt.Sprintf("(%s >= %s)", g.Params[0].Expr(), g.Params[1].Expr())
}

func (g *GreaterThanOrEqual) RealizedExpr(eCtx ast_ifaces.Context) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s >= %s)", g.Params[0].RealizedExpr(eCtx), g.Params[1].RealizedExpr(eCtx))
}

func (g *GreaterThanOrEqual) GetCtn() ast_ifaces.Container {
	return g.Ctn
}

func (g *GreaterThanOrEqual) SetCtn(c ast_ifaces.Container) {
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
