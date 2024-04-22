package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type FromJsonFn struct {
	Fn
}

func (a *FromJsonFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitFromJsonFn(eCtx, a)
}

func (f *FromJsonFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
	json := evaluator.EvaluateWithContext(eCtx, f.Parameters()[0]).ConvertToString()
	// TODO: implement real logic with PipelineContextData
	// return runner.ToPipelineContextData(json)
	return json
}

func (f *FromJsonFn) TraceFullyRealized() bool {
	return false
}

func (f *FromJsonFn) SetName(name string) {
	f.Name = name
}

func (f *FromJsonFn) GetName() string {
	return f.Name
}

func (f *FromJsonFn) GetContainer() interfaces.IContainer {
	return f.Container
}

func (f *FromJsonFn) SetContainer(c interfaces.IContainer) {
	f.Container = c
}

func (f *FromJsonFn) ConvertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *FromJsonFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}
