package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type LessThanOrEqual struct {
	base.ExpressionNodeBase
	base.ContainerBase
}

func (l *LessThanOrEqual) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitLessThanOrEqual(eCtx, l)
}

func (l *LessThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (l *LessThanOrEqual) ConvertToExpression() string {
	return fmt.Sprintf("(%s <= %s)", l.Params[0].ConvertToExpression(), l.Params[1].ConvertToExpression())
}

func (l *LessThanOrEqual) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s <= %s)", l.Params[0].ConvertToRealizedExpression(eCtx), l.Params[1].ConvertToRealizedExpression(eCtx))
}

func (l *LessThanOrEqual) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	left := evaluator.EvaluateWithContext(eCtx, l.Params[0])
	right := evaluator.EvaluateWithContext(eCtx, l.Params[1])
	return left.AbstractLessThanOrEqual(right)
}

func (l *LessThanOrEqual) GetContainer() interfaces.IContainer {
	return l.Container
}

func (l *LessThanOrEqual) SetContainer(c interfaces.IContainer) {
	l.Container = c
}

func (l *LessThanOrEqual) GetLevel() (level int) {
	return l.Level
}

func (l *LessThanOrEqual) GetName() string {
	return l.Name
}

func (l *LessThanOrEqual) SetName(name string) {
	l.Name = name
}
