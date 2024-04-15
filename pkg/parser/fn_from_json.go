package parser

import (
	"fmt"
	"strings"
)

type FromJsonFn struct {
	Fn
}

func (f *FromJsonFn) evaluateCore(eCtx *EvaluationContext) any {
	json := evaluate(eCtx, f.Parameters()[0]).ConvertToString()
	// TODO: implement real logic with PipelineContextData
	// return runner.ToPipelineContextData(json)
	return json
}

func (f *FromJsonFn) traceFullyRealized() bool {
	return false
}

func (f *FromJsonFn) setName(name string) {
	f.name = name
}

func (f *FromJsonFn) getName() string {
	return f.name
}

func (f *FromJsonFn) getContainer() iContainer {
	return f.container
}

func (f *FromJsonFn) setContainer(c iContainer) {
	f.container = c
}

func (f *FromJsonFn) convertToExpression() string {
	params := make([]string, len(f.Parameters()))
	for i, param := range f.Parameters() {
		params[i] = param.convertToExpression()
	}
	return fmt.Sprintf("%s(%s)", f.getName(), strings.Join(params, ", "))
}

func (f *FromJsonFn) convertToRealizedExpression(eCtx *EvaluationContext) string {
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
