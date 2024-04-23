package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type NotEqual struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (n *NotEqual) Value() any {
	panic("not implemented")
}

func (n *NotEqual) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitNotEqual(eCtx, n)
}

func (n *NotEqual) TraceFullyRealized() bool {
	return false
}

func (n *NotEqual) ConvertToExpression() string {
	return fmt.Sprintf("%s != %s", n.Params[0].ConvertToExpression(), n.Params[1].ConvertToExpression())
}

func (n *NotEqual) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return fmt.Sprintf("%s != %s", n.Params[0].ConvertToRealizedExpression(eCtx), n.Params[1].ConvertToRealizedExpression(eCtx))
}

func (n *NotEqual) GetContainer() expression.IContainer {
	return n.Container
}

func (n *NotEqual) SetContainer(c expression.IContainer) {
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
