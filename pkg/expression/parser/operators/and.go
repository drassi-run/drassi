package operators

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type And struct {
	base.ContainerBs
	base.ExpressionNodeBs
}

func (a *And) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
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

func (a *And) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
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

//
// func (a *And) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	result := &evaluator.EvaluationResult{}
// 	for _, param := range a.Params {
// 		result = evaluator.EvaluateWithContext(eCtx, param)
// 		if result.IsFalsy() {
// 			return result.Value()
// 		}
// 	}
// 	return result.Value()
// }

func (a *And) TraceFullyRealized() bool {
	return false
}

func (a *And) GetContainer() interfaces.IContainer {
	return a.Container
}

func (a *And) SetContainer(c interfaces.IContainer) {
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
