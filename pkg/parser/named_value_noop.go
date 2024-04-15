package parser

type NoOpNamedValue struct {
	NamedValue
}

func (c *NoOpNamedValue) evaluateCore(eCtx *EvaluationContext) (result any) {

	return nil
}

func (c *NoOpNamedValue) traceFullyRealized() bool {
	return true
}

func (c *NoOpNamedValue) getContainer() iContainer {
	return c.container
}

func (c *NoOpNamedValue) setContainer(cc iContainer) {
	c.container = cc
}

func (c *NoOpNamedValue) setName(name string) {
	c.ExpressionNode.name = name
}

func (c *NoOpNamedValue) getName() string {
	return c.name
}
