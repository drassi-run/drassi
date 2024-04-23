package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type GreaterThan struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (g *GreaterThan) Value() any {
	panic("not implemented")
}

func (g *GreaterThan) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitGreaterThan(eCtx, g)
}

func (g *GreaterThan) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThan) ConvertToExpression() string {
	return fmt.Sprintf("(%s > %s)", g.Params[0].ConvertToExpression(), g.Params[1].ConvertToExpression())
}

func (g *GreaterThan) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s > %s)", g.Params[0].ConvertToRealizedExpression(eCtx), g.Params[1].ConvertToRealizedExpression(eCtx))
}

func (g *GreaterThan) GetContainer() expression.IContainer {
	return g.Container
}

func (g *GreaterThan) SetContainer(c expression.IContainer) {
	g.Container = c
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
