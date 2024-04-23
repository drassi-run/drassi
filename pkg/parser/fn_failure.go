package parser

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/runner"
)

type FailureFn struct {
	Fn
}

func (f *FailureFn) evaluateCore(eCtx *EvaluationContext) any {
	tplCtx := eCtx.State().(*runner.TemplateContext)
	if tplCtx == nil {
		panic(ErrorsTemplateContextNotFound)
	}
	// TODO: refactor me
	execCtx := tplCtx.State["IExecutionContext"].(runner.IExecutionContext)
	if execCtx == nil {
		panic(ErrorsExecutionContextNotFound)
	}
	// Decide based on 'action_status' for composite MAIN steps and 'job.status' for pre, post and job-getLevel steps
	isComposite := execCtx.IsEmbedded() && execCtx.Stage() == runner.ActionRunStageMain
	if isComposite {
		// TODO: refactor me
		// If status is not parseable, evaluate actionStatus to ActionResultSuccess
		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
			if ok := runner.TryParseActionResult(actionStatus); ok {
				return actionStatus == fmt.Sprintf("%s", runner.ActionResultFailure)
			}
			return false
		}
	}
	return execCtx.JobContext().Status == runner.ActionResultFailure
}

func (f *FailureFn) setName(name string) {
	f.name = name
}

func (f *FailureFn) getName() string {
	return f.name
}

func (f *FailureFn) getContainer() iContainer {
	return f.container
}

func (f *FailureFn) setContainer(cc iContainer) {
	f.container = cc
}

func (f *FailureFn) traceFullyRealized() bool {
	return false
}

func (f *FailureFn) convertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}

func (f *FailureFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(f)
	if exist {
		return result
	}
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}
