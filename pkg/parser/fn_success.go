package parser

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/runner"
)

type SuccessFn struct {
	Fn
}

func (s *SuccessFn) evaluateCore(eCtx *EvaluationContext) any {
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
		// If status is not parsable, evaluate actionStatus to ActionResultSuccess
		if actionStatus := execCtx.GetGitHubContext("action_status"); actionStatus != "" {
			if ok := runner.TryParseActionResult(actionStatus); ok {
				return actionStatus == fmt.Sprintf("%s", runner.ActionResultSuccess)
			}
			return true
		}
	}
	return execCtx.JobContext().Status == runner.ActionResultSuccess
}

func (s *SuccessFn) setName(name string) {
	s.name = name
}

func (s *SuccessFn) getName() string {
	return s.name
}

func (s *SuccessFn) getContainer() iContainer {
	return s.container
}

func (s *SuccessFn) setContainer(cc iContainer) {
	s.container = cc
}

func (s *SuccessFn) traceFullyRealized() bool {
	return false
}

func (s *SuccessFn) convertToExpression() string {
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", s.getName(), strings.Join(params, ", "))
}

func (s *SuccessFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
	exist, result := eCtx.tryGetTraceResult(s)
	if exist {
		return result
	}
	params := make([]string, len(s.Parameters()))
	for i, param := range s.Parameters() {
		params[i] = param.convertToRealizedExpression(eCtx)
	}
	return fmt.Sprintf("%s(%s)", s.getName(), strings.Join(params, ", "))
}
