package operators

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type And struct {
	base.ContainerBs
	base.ExpressionNodeBs
}

func (a *And) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitAnd(eCtx, a)
}

func (a *And) Value() any {
	panic("not implemented")
}

func (a *And) ConvertToExpression() string {
	expressions := make([]string, len(a.Params))
	for i, param := range a.Params {
		expressions[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(a)
	if exist {
		return result
	}
	expressions := make([]string, len(a.Params))
	for i, param := range a.Params {
		expressions[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " && "))
}

func (a *And) TraceFullyRealized() bool {
	return false
}

func (a *And) GetContainer() expression.IContainer {
	return a.Container
}

func (a *And) SetContainer(c expression.IContainer) {
	a.Container = c
}

func (a *And) GetLevel() (level int) {
	return a.Level
}

func (a *And) GetName() string {
	return a.Name
}

func (a *And) SetName(name string) {
	a.Name = name
}
