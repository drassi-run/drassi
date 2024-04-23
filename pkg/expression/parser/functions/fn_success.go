package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type SuccessFn struct {
	Fn
}

func (s *SuccessFn) Value() any {
	panic("not implemented")
}

func (s *SuccessFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitSuccessFn(eCtx, s)
}
func (s *SuccessFn) SetName(name string) {
	s.Name = name
}

func (s *SuccessFn) GetName() string {
	return s.Name
}

func (s *SuccessFn) GetContainer() interfaces.IContainer {
	return s.Container
}

func (s *SuccessFn) SetContainer(cc interfaces.IContainer) {
	s.Container = cc
}

func (s *SuccessFn) TraceFullyRealized() bool {
	return false
}

func (s *SuccessFn) ConvertToExpression() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.ConvertToExpression()
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}

func (s *SuccessFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
	exist, result := eCtx.TryGetTraceResult(s)
	if exist {
		return result
	}
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.ConvertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", s.GetName(), strings.Join(params, ", "))
}
