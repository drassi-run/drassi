package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type CancelledFn struct {
	Fn
}

var (
	ErrorsTemplateContextNotFound  = "template context not found"
	ErrorsExecutionContextNotFound = "execution context not found"
)

func (c *CancelledFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpressionNodeVisitor) any {
	return v.VisitCancelledFn(eCtx, c)
}

func (c *CancelledFn) Value() any {
	panic("not implemented")
}

func (c *CancelledFn) ConvertToExpression() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *CancelledFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(c)
	if exist {
		return result
	}
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *CancelledFn) SetName(name string) {
	c.Name = name
}

func (c *CancelledFn) GetName() string {
	return c.Name
}

func (c *CancelledFn) GetContainer() expression.IContainer {
	return c.Container
}

func (c *CancelledFn) SetContainer(cc expression.IContainer) {
	c.Container = cc
}

func (c *CancelledFn) TraceFullyRealized() bool {
	return false
}
