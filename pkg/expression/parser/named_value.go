package parser

import (
	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type INamedValue interface {
	expression.IExpNode
}

// NamedValue represent keyword available from parseContext eg: github, job, steps, env
type NamedValue struct {
	INamedValue
	base.ExpressionNodeBs
}

func (n *NamedValue) ConvertToExpression() string {
	return n.Name
}

func (n *NamedValue) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return n.Name
}

func (n *NamedValue) GetLevel() (level int) {
	return n.Level
}
