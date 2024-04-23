package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type LessThanOrEqual struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (l *LessThanOrEqual) Value() any {
	panic("not implemented")
}

func (l *LessThanOrEqual) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitLessThanOrEqual(eCtx, l)
}

func (l *LessThanOrEqual) TraceFullyRealized() bool {
	return false
}

func (l *LessThanOrEqual) ConvertToExpression() string {
	return fmt.Sprintf("(%s <= %s)", l.Params[0].ConvertToExpression(), l.Params[1].ConvertToExpression())
}

func (l *LessThanOrEqual) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s <= %s)", l.Params[0].ConvertToRealizedExpression(eCtx), l.Params[1].ConvertToRealizedExpression(eCtx))
}

func (l *LessThanOrEqual) GetContainer() expression.IContainer {
	return l.Container
}

func (l *LessThanOrEqual) SetContainer(c expression.IContainer) {
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
