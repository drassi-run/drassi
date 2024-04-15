package parser

import (
	"expression_parser/runner"
)

type ContextValueNode struct {
	NamedValue
	ExpressionNode
}

func (c *ContextValueNode) evaluateCore(eCtx *EvaluationContext) any {
	return eCtx.State().(*runner.TemplateContext).ExpressionValues[c.ExpressionNode.name]
}

func (c *ContextValueNode) traceFullyRealized() bool {
	return true
}

func (c *ContextValueNode) getContainer() iContainer {
	return c.container
}

func (c *ContextValueNode) setContainer(cc iContainer) {
	c.container = cc
}

func (c *ContextValueNode) setName(name string) {
	c.ExpressionNode.name = name
}

func (c *ContextValueNode) getName() string {
	return c.name
}
