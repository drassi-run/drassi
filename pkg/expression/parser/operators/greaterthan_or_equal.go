package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type GreaterThanOrEqual struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (a *GreaterThanOrEqual) Value() any {
	panic("not implemented")
}

func (a *GreaterThanOrEqual) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitGreaterThanOrEqual(eCtx, a)
}

func (g *GreaterThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThanOrEqual) ConvertToExpression() string {
	return fmt.Sprintf("(%s >= %s)", g.Params[0].ConvertToExpression(), g.Params[1].ConvertToExpression())
}

func (g *GreaterThanOrEqual) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s >= %s)", g.Params[0].ConvertToRealizedExpression(eCtx), g.Params[1].ConvertToRealizedExpression(eCtx))
}

// func (g *GreaterThanOrEqual) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, g.Params[0])
// 	r := evaluator.EvaluateWithContext(eCtx, g.Params[1])
// 	return l.AbstractGreaterThan(r)
// }

func (g *GreaterThanOrEqual) GetContainer() interfaces.IContainer {
	return g.Container
}

func (g *GreaterThanOrEqual) SetContainer(c interfaces.IContainer) {
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
