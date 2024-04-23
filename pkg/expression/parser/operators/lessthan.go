package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type LessThan struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (l *LessThan) Value() any {
	panic("not implemented")
}

func (l *LessThan) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitLessThan(eCtx, l)
}

func (l *LessThan) TraceFullyRealized() bool {
	return false
}

func (l *LessThan) ConvertToExpression() string {
	return fmt.Sprintf("(%s < %s)", l.Params[0].ConvertToExpression(), l.Params[1].ConvertToExpression())
}

func (l *LessThan) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(l)
	if exist {
		return result
	}
	return fmt.Sprintf("(%s < %s)", l.Params[0].ConvertToRealizedExpression(eCtx), l.Params[1].ConvertToRealizedExpression(eCtx))
}

func (l *LessThan) GetContainer() interfaces.IContainer {
	return l.Container
}

func (l *LessThan) SetContainer(c interfaces.IContainer) {
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
