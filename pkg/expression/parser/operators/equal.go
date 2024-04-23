package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Equal struct {
	base.ContainerBs
	base.ExpressionNodeBs
}

func (e *Equal) Value() any {
	panic("not implemented")
}

func (e *Equal) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitEqual(eCtx, e)
}

func (e *Equal) TraceFullyRealized() bool {
	return false
}

func (e *Equal) ConvertToExpression() string {
	return fmt.Sprintf("(%s == %s)", e.Params[0].ConvertToExpression(), e.Params[1].ConvertToExpression())
}

func (e *Equal) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(e)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s == %s)", e.Params[0].ConvertToRealizedExpression(eCtx), e.Params[1].ConvertToRealizedExpression(eCtx))
}

// func (e *Equal) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	l := evaluator.EvaluateWithContext(eCtx, e.Params[0])
// 	r := evaluator.EvaluateWithContext(eCtx, e.Params[1])
// 	return l.AbstractEqual(r)
// }

func (e *Equal) GetContainer() interfaces.IContainer {
	return e.Container
}

func (e *Equal) SetContainer(c interfaces.IContainer) {
	e.Container = c
}

func (e *Equal) GetLevel() (level int) {
	return e.Level
}

func (e *Equal) GetName() string {
	return e.Name
}
func (e *Equal) SetName(name string) {
	e.Name = name
}
