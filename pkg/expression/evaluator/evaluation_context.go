package evaluator

import (
	"errors"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type (
	evaluationContext struct {
		state        any
		options      *EvaluationOption
		trace        interfaces.ITraceWriter
		masker       interfaces.ISecretMasker
		traceResults map[interfaces.IExpressionNode]string
	}
)

var (
	ErrorsEmptyTrace        = errors.New("trace must be provided")
	ErrorsEmptySecretMasker = errors.New("secret masker must be provider")
)

func (e *evaluationContext) State() any {
	return e.state
}

func (e *evaluationContext) Masker() interfaces.ISecretMasker {
	return e.masker
}

func (e *evaluationContext) Trace() interfaces.ITraceWriter {
	return e.trace
}

func newEvaluationContext(trace interfaces.ITraceWriter, masker interfaces.ISecretMasker, state any, options *EvaluationOption,
	node interfaces.IExpressionNode) *evaluationContext {
	if trace == nil {
		panic(ErrorsEmptyTrace)
	}
	if masker == nil {
		panic(ErrorsEmptySecretMasker)
	}
	e := &evaluationContext{
		state:  state,
		trace:  trace,
		masker: masker,
	}
	e.options = options
	e.traceResults = map[interfaces.IExpressionNode]string{}
	return e
}

func (e *evaluationContext) SetTraceResult(node interfaces.IExpressionNode, result interfaces.IEvaluationResult) {
	if _, exist := e.traceResults[node]; exist {
		delete(e.traceResults, node)
	}
	value := formatValueFromResult(e.masker, result)
	e.traceResults[node] = value
}

func (e *evaluationContext) TryGetTraceResult(node interfaces.IExpressionNode) (exist bool, result string) {
	result, exist = e.traceResults[node]
	return
}
