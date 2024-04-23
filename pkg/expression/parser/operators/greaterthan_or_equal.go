package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type GreaterThanOrEqual struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (g *GreaterThanOrEqual) Value() any {
	panic("not implemented")
}

func (g *GreaterThanOrEqual) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitGreaterThanOrEqual(eCtx, g)
}

func (g *GreaterThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThanOrEqual) ConvertToExpression() string {
	return fmt.Sprintf("(%s >= %s)", g.Params[0].ConvertToExpression(), g.Params[1].ConvertToExpression())
}

func (g *GreaterThanOrEqual) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s >= %s)", g.Params[0].ConvertToRealizedExpression(eCtx), g.Params[1].ConvertToRealizedExpression(eCtx))
}

func (g *GreaterThanOrEqual) GetContainer() expression.IContainer {
	return g.Container
}

func (g *GreaterThanOrEqual) SetContainer(c expression.IContainer) {
	g.Container = c
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
