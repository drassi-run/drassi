package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Not struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (a *Not) Value() any {
	panic("not implemented")
}

func (a *Not) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitNot(eCtx, a)
}

func (n *Not) TraceFullyRealized() bool {
	return false
}

func (n *Not) ConvertToExpression() string {
	return fmt.Sprintf("!%s", n.Params[0].ConvertToExpression())
}

func (n *Not) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("!%s", n.Params[0].ConvertToRealizedExpression(eCtx))
}

//
// func (n *Not) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	return evaluator.EvaluateWithContext(eCtx, n.Params[0]).IsFalsy()
// }

func (n *Not) GetContainer() interfaces.IContainer {
	return n.Container
}

func (n *Not) SetContainer(c interfaces.IContainer) {
	n.Container = c
}

func (n *Not) GetLevel() (level int) {
	return n.Level
}

func (n *Not) GetName() string {
	return n.Name
}

func (n *Not) SetName(name string) {
	n.Name = name
}
