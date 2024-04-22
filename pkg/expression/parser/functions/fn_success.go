package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type SuccessFn struct {
	Fn
}

func (a *SuccessFn) Value() any {
	panic("not implemented")
}

func (s *SuccessFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitSuccessFn(eCtx, s)
}

//
// func (s *SuccessFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
// 	tplCtx := eCtx.State().(*runner.TemplateContext)
// 	if tplCtx == nil {
// 		panic(ErrorsTemplateContextNotFound)
// 	}
// 	// TODO: refactor me
// 	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
// 	if execCtx == nil {
// 		panic(ErrorsExecutionContextNotFound)
// 	}
// 	// Decide based on 'action_status' for composite MAIN steps and 'job.status' for pre, post and job-getLevel steps
// 	isComposite := execCtx.IsEmbedded() && execCtx.Stage() == runner.ActionRunStageMain
// 	if isComposite {
// 		// TODO: refactor me
// 		// If status is not parsable, evaluate actionStatus to ActionResultSuccess
// 		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
// 			if ok := runner.TryParseActionResult(actionStatus); ok {
// 				return actionStatus == fmt.Sprintf("%s", runner.ActionResultSuccess)
// 			}
// 			return true
// 		}
// 	}
// 	return execCtx.JobContext().Status == runner.ActionResultSuccess
// }

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
