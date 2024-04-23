package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type NoOpFn struct {
	base.ExpressionNodeBs
	Fn
}

func (n *NoOpFn) Value() any {
	panic("not implemented")
}

func (n *NoOpFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitNoopFn(eCtx, n)
}

// func (n *NoOpFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	return nil
// }

func (n *NoOpFn) TraceFullyRealized() bool {
	return false
}

func (n *NoOpFn) GetContainer() expression.IContainer {
	return n.Container
}

func (n *NoOpFn) SetContainer(c expression.IContainer) {
	n.Container = c
}

func (n *NoOpFn) GetLevel() (level int) {
	return n.Level
}

func (n *NoOpFn) GetName() string {
	return n.Name
}
func (n *NoOpFn) SetName(name string) {
	n.Name = name
}

func (n *NoOpFn) ConvertToExpression() string {
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", n.GetName(), strings.Join(params, ", "))
}

func (n *NoOpFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(n)
	if exist {
		return result
	}
	params := make([]string, len(n.Parameters()))
	for i, param := range n.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", n.GetName(), strings.Join(params, ", "))
}
