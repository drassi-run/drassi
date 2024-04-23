package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type LessThan struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (l *LessThan) Value() any {
	panic("not implemented")
}

func (l *LessThan) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitLessThan(eCtx, l)
}

func (l *LessThan) TraceFullyRealized() bool {
	return false
}

func (l *LessThan) ConvertToExpression() string {
	return fmt.Sprintf("(%s < %s)", l.Params[0].ConvertToExpression(), l.Params[1].ConvertToExpression())
}

func (l *LessThan) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s < %s)", l.Params[0].ConvertToRealizedExpression(eCtx), l.Params[1].ConvertToRealizedExpression(eCtx))
}

func (l *LessThan) GetContainer() expression.IContainer {
	return l.Container
}

func (l *LessThan) SetContainer(c expression.IContainer) {
	l.Container = c
}

func (l *LessThan) GetLevel() (level int) {
	return l.Level
}

func (l *LessThan) GetName() string {
	return l.Name
}

func (l *LessThan) SetName(name string) {
	l.Name = name
}
