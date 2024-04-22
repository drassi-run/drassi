package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/runner"
)

type CancelledFn struct {
	Fn
}

var (
	ErrorsTemplateContextNotFound  = "template context not found"
	ErrorsExecutionContextNotFound = "execution context not found"
)

func (a *CancelledFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitCancelledFn(eCtx, a)
}

func (c *CancelledFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	return execCtx.JobContext().Status == runner.ActionResultCancelled
}

func (c *CancelledFn) ConvertToExpression() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", c.GetName(), strings.Join(params, ", "))
}

func (c *CancelledFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
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

func (c *CancelledFn) GetContainer() interfaces.IContainer {
	return c.Container
}

func (c *CancelledFn) setContainer(cc interfaces.IContainer) {
	c.Container = cc
}

func (c *CancelledFn) TraceFullyRealized() bool {
	return false
}
