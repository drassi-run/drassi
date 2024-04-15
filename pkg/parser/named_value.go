package parser

type iNamedValue interface {
	IExpressionNode
}

// NamedValue represent keyword available from parseContext eg: github, job, steps, env
type NamedValue struct {
	iNamedValue
	ExpressionNode
}

func (n *NamedValue) convertToExpression() string {
	return n.name
}

func (n *NamedValue) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(n)
	if exist {
		return result
	}
	return n.name
}

func (n *NamedValue) getLevel() (level int) {
	return n.level
}
