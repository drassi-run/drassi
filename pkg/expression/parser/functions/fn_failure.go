package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type FailureFn struct {
	Fn
}

func (f *FailureFn) Value() any {
	panic("not implemented")
}

func (f *FailureFn) Accept(eCtx expression.IEvaluationContext, v expression.IExpNodeVisitor) any {
	return v.VisitFailureFn(eCtx, f)
}

func (f *FailureFn) SetName(name string) {
	f.Name = name
}

func (f *FailureFn) GetName() string {
	return f.Name
}

func (f *FailureFn) GetContainer() expression.IContainer {
	return f.Container
}

func (f *FailureFn) SetContainer(cc expression.IContainer) {
	f.Container = cc
}

func (f *FailureFn) TraceFullyRealized() bool {
	return false
}

func (f *FailureFn) ConvertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.GetName(), strings.Join(params, ", "))
}

func (f *FailureFn) ConvertToRealizedExpression(eCtx expression.IEvaluationContext) string {
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
