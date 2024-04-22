package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type GreaterThan struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (a *GreaterThan) Value() any {
	panic("not implemented")
}

func (a *GreaterThan) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitGreaterThan(eCtx, a)
}

func (g *GreaterThan) TraceFullyRealized() bool {
	return false
}

func (g *GreaterThan) ConvertToExpression() string {
	return fmt.Sprintf("(%s > %s)", g.Params[0].ConvertToExpression(), g.Params[1].ConvertToExpression())
}

func (g *GreaterThan) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(g)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s > %s)", g.Params[0].ConvertToRealizedExpression(eCtx), g.Params[1].ConvertToRealizedExpression(eCtx))
}

//
// func (g *GreaterThan) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, g.Params[0])
// 	r := evaluator.EvaluateWithContext(eCtx, g.Params[1])
// 	return l.AbstractGreaterThan(r)
// }

func (g *GreaterThan) GetContainer() interfaces.IContainer {
	return g.Container
}

func (g *GreaterThan) SetContainer(c interfaces.IContainer) {
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
