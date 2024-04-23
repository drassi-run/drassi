package evaluator

import (
	"errors"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type (
	evaluationContext struct {
		state        any
		options      *EvaluationOption
		trace        expression.ITraceWriter
		masker       interfaces.ISecretMasker
		traceResults map[expression.IExpressionNode]string
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

func (e *evaluationContext) Trace() expression.ITraceWriter {
	return e.trace
}

func newEvaluationContext(trace expression.ITraceWriter, masker interfaces.ISecretMasker, state any, options *EvaluationOption,
	node expression.IExpressionNode) *evaluationContext {
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
	e.traceResults = map[expression.IExpressionNode]string{}
	return e
}

func (e *evaluationContext) SetTraceResult(node expression.IExpressionNode, result expression.IEvaluationResult) {
	if _, exist := e.traceResults[node]; exist {
		delete(e.traceResults, node)
	}
	value := formatValueFromResult(e.masker, result)
	e.traceResults[node] = value
}

func (e *evaluationContext) TryGetTraceResult(node expression.IExpressionNode) (exist bool, result string) {
	result, exist = e.traceResults[node]
	return
}
