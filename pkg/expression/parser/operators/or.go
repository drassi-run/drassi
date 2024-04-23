package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Or struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (o *Or) Value() any {
	panic("not implemented")
}

func (o *Or) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitOr(eCtx, o)
}

func (o *Or) TraceFullyRealized() bool {
	return false
}

func (o *Or) ConvertToExpression() string {
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToExpression(), o.Params[1].ConvertToExpression())
}

func (o *Or) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToRealizedExpression(eCtx), o.Params[1].ConvertToRealizedExpression(eCtx))
}

func (o *Or) GetContainer() expression.IContainer {
	return o.Container
}

func (o *Or) SetContainer(c expression.IContainer) {
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
