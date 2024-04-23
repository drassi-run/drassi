package parser

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/runner"
)

type CancelledFn struct {
	Fn
}

var (
	ErrorsTemplateContextNotFound  = "template context not found"
	ErrorsExecutionContextNotFound = "execution context not found"
)

func (c *CancelledFn) evaluateCore(eCtx *EvaluationContext) any {
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

func (c *CancelledFn) convertToExpression() string {
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", c.getName(), strings.Join(params, ", "))
}

func (c *CancelledFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(c)
	if exist {
		return result
	}
	params := make([]string, len(c.Parameters()))
	for i, param := range c.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", c.getName(), strings.Join(params, ", "))
}

func (c *CancelledFn) setName(name string) {
	c.name = name
}

func (c *CancelledFn) getName() string {
	return c.name
}

func (c *CancelledFn) getContainer() iContainer {
	return c.container
}

func (c *CancelledFn) setContainer(cc iContainer) {
	c.container = cc
}

func (c *CancelledFn) traceFullyRealized() bool {
	return false
}
