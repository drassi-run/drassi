package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Not struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (n *Not) Value() any {
	panic("not implemented")
}

func (n *Not) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitNot(eCtx, n)
}

func (n *Not) TraceFullyRealized() bool {
	return false
}

func (n *Not) ConvertToExpression() string {
	return fmt.Sprintf("!%s", n.Params[0].ConvertToExpression())
}

func (n *Not) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("!%s", n.Params[0].ConvertToRealizedExpression(eCtx))
}

func (n *Not) GetContainer() expression.IContainer {
	return n.Container
}

func (n *Not) SetContainer(c expression.IContainer) {
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
