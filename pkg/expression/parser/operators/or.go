package operators

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/base"
)

type Or struct {
	base.ExpressionNodeBs
	base.ContainerBs
}

func (a *Or) Value() any {
	panic("not implemented")
}

func (a *Or) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitOr(eCtx, a)
}

func (o *Or) TraceFullyRealized() bool {
	return false
}

func (o *Or) ConvertToExpression() string {
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToExpression(), o.Params[1].ConvertToExpression())
}

func (o *Or) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(o)
	if exist {
		return result
	}
	return fmt.Sprintf("%s || %s", o.Params[0].ConvertToRealizedExpression(eCtx), o.Params[1].ConvertToRealizedExpression(eCtx))
}

//
// func (o *Or) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	var result *evaluator.EvaluationResult
// 	for _, param := range o.Params {
// 		result = evaluator.EvaluateWithContext(eCtx, param)
// 		if result.IsTruthy() {
// 			break
// 		}
// 	}
// 	if result == nil {
// 		return nil
// 	}
// 	return result.Value()
// }

func (o *Or) GetContainer() interfaces.IContainer {
	return o.Container
}

func (o *Or) SetContainer(c interfaces.IContainer) {
	o.Container = c
}

func (o *Or) GetLevel() (level int) {
	return o.Level
}

func (o *Or) GetName() string {
	return o.Name
}

func (o *Or) SetName(name string) {
	o.Name = name
}
