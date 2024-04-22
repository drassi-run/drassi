package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type INamedValue interface {
	interfaces.IExpressionNode
}

// NamedValue represent keyword available from parseContext eg: github, job, steps, env
type NamedValue struct {
	INamedValue
	base.ExpressionNodeBs
}

func (n *NamedValue) ConvertToExpression() string {
	return n.Name
}

func (n *NamedValue) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	return n.Name
}

func (n *NamedValue) GetLevel() (level int) {
	return n.Level
}
