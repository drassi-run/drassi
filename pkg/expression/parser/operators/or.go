package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Or struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (o *Or) Value() any {
	panic("not implemented")
}

func (o *Or) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitOr(eCtx, o)
}

func (o *Or) TraceFullyRealized() bool {
	return false
}

func (o *Or) ConvertToExpression() string {
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToExpression(), o.Params[1].ConvertToExpression())
}

func (o *Or) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToRealizedExpression(eCtx), o.Params[1].ConvertToRealizedExpression(eCtx))
}

func (o *Or) GetContainer() interfaces.IContainer {
	return o.Container
}

func (o *Or) SetContainer(c interfaces.IContainer) {
	o.Container = c
}

func (o *Or) GetLevel() (level int) {
	return o.Level
}

func (o *Or) GetName() string {
	return o.Name
}

func (o *Or) SetName(name string) {
	o.Name = name
}
