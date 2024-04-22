package functions

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type FailureFn struct {
	Fn
}

func (a *FailureFn) Value() any {
	panic("not implemented")
}

func (a *FailureFn) Accept(eCtx interfaces.IEvaluationContext, v interfaces.IExpressionNodeVisitor) any {
	return v.VisitFailureFn(eCtx, a)
}

//
// func (f *FailureFn) EvaluateCore(eCtx interfaces.IEvaluationContext) any {
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
// 		// If status is not parseable, evaluate actionStatus to ActionResultSuccess
// 		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
// 			if ok := runner.TryParseActionResult(actionStatus); ok {
// 				return actionStatus == fmt.Sprintf("%s", runner.ActionResultFailure)
// 			}
// 			return false
// 		}
// 	}
// 	return execCtx.JobContext().Status == runner.ActionResultFailure
// }

func (f *FailureFn) SetName(name string) {
	f.Name = name
}

func (f *FailureFn) GetName() string {
	return f.Name
}

func (f *FailureFn) GetContainer() interfaces.IContainer {
	return f.Container
}

func (f *FailureFn) SetContainer(cc interfaces.IContainer) {
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

func (f *FailureFn) ConvertToRealizedExpression(eCtx interfaces.IEvaluationContext) string {
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
