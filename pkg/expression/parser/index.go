package parser

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type (
	index struct {
		base.ExpressionNodeBs
		base.ContainerBs
	}
)

func (i *index) Value() any {
	panic("not implemented")
}

func (i *index) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitIndex(eCtx, i)
}

func (i *index) TraceFullyRealized() bool {
	return true
}

func (i *index) ConvertToExpression() string {
	// Verify if we can simplify the expression, we would rather return
	// github.sha then github['sha'] so we check if this is a simple case.
	if lt, ok := i.Params[1].(*literal); ok {
		if lStr, ok := lt.Value().(string); ok && isLegalKeyWord(lStr) {
			return fmt.Sprintf("%s.%s", i.Params[0].ConvertToExpression(), i.Params[0].ConvertToExpression())
		}
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].ConvertToExpression(), i.Params[1].ConvertToExpression())
}

func (i *index) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(i)
	if exist {
		return result
	}
	return fmt.Sprintf("%s[%s]", i.Params[0].ConvertToExpression(), i.Params[1].ConvertToExpression())
}

func (i *index) GetContainer() interfaces.IContainer {
	return i.Container
}

func (i *index) SetContainer(c interfaces.IContainer) {
	i.Container = c
}

func (i *index) GetLevel() (level int) {
	return i.Level
}

func (i *index) GetName() string {
	return i.Name
}

func (i *index) SetName(name string) {
	i.Name = name
}
