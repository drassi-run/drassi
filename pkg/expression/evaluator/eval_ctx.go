package evaluator

import (
	"errors"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type (
	EvaluationContext struct {
		state        any
		options      *EvaluationOption
		trace        interfaces.ITraceWriter
		masker       interfaces.ISecretMasker
		traceResults map[interfaces.IExpressionNode]string
	}
)

const (
	OneMegaBytes = 1048576
)

var (
	ErrorsEmptyTrace        = errors.New("trace must be provided")
	ErrorsEmptySecretMasker = errors.New("secret masker must be provider")
)

func (e *EvaluationContext) State() any {
	return e.state
}

func (e *EvaluationContext) Masker() interfaces.ISecretMasker {
	return e.masker
}

func (e *EvaluationContext) Trace() interfaces.ITraceWriter {
	return e.trace
}

func NewEvaluationContext(trace interfaces.ITraceWriter, masker interfaces.ISecretMasker, state any, options *EvaluationOption,
	node interfaces.IExpressionNode) *EvaluationContext {
	if trace == nil {
		panic(ErrorsEmptyTrace)
	}
	if masker == nil {
		panic(ErrorsEmptySecretMasker)
	}
	e := &EvaluationContext{
		state:  state,
		trace:  trace,
		masker: masker,
	}
	e.options = options
	e.traceResults = map[interfaces.IExpressionNode]string{}
	return e
}

func (e *EvaluationContext) SetTraceResult(node interfaces.IExpressionNode, result interfaces.IEvaluationResult) {
	if _, exist := e.traceResults[node]; exist {
		delete(e.traceResults, node)
	}
	value := FormatValueFromResult(e.masker, result)
	e.traceResults[node] = value
}

func (e *EvaluationContext) TryGetTraceResult(node interfaces.IExpressionNode) (exist bool, result string) {
	result, exist = e.traceResults[node]
	return
}
