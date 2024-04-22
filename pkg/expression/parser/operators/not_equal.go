package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type NotEqual struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (a *NotEqual) Value() any {
	panic("not implemented")
}

func (a *NotEqual) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitNotEqual(eCtx, a)
}

func (n *NotEqual) TraceFullyRealized() bool {
	return false
}

func (n *NotEqual) ConvertToExpression() string {
	return fmt.Sprintf("%s != %s", n.Params[0].ConvertToExpression(), n.Params[1].ConvertToExpression())
}

func (n *NotEqual) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("%s != %s", n.Params[0].ConvertToRealizedExpression(eCtx), n.Params[1].ConvertToRealizedExpression(eCtx))
}

//
// func (n *NotEqual) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, n.Params[0])
// 	r := evaluator.EvaluateWithContext(eCtx, n.Params[1])
// 	return l.AbstractNotEqual(r)
// }

func (n *NotEqual) GetContainer() interfaces.IContainer {
	return n.Container
}

func (n *NotEqual) SetContainer(c interfaces.IContainer) {
	n.Container = c
}

func (n *NotEqual) GetLevel() (level int) {
	return n.Level
}

func (n *NotEqual) GetName() string {
	return n.Name
}

func (n *NotEqual) SetName(name string) {
	n.Name = name
}
