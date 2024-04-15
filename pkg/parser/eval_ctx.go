package parser

import (
	"errors"
)

type (
	EvaluationContext struct {
		state        any
		memory       *evaluationMemory
		options      *EvaluationOption
		trace        ITraceWriter
		masker       ISecretMasker
		traceResults map[IExpressionNode]string
		traceMemory  *MemoryCounter
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

func (e *EvaluationContext) Masker() ISecretMasker {
	return e.masker
}

func (e *EvaluationContext) Trace() ITraceWriter {
	return e.trace
}

func NewEvaluationContext(trace ITraceWriter, masker ISecretMasker, state any, options *EvaluationOption,
	node IExpressionNode) *EvaluationContext {
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
	opt := *options
	if opt.GetMaxMemory() == 0 {
		opt.SetMaxMemory(OneMegaBytes)
	}
	e.options = &opt
	e.memory = newEvaluationMemory(opt.GetMaxMemory(), node)
	e.traceResults = map[IExpressionNode]string{}
	e.traceMemory = NewMemoryCounter(nil, opt.GetMaxMemory())
	return e
}

func (e *EvaluationContext) setTraceResult(node IExpressionNode, result *EvaluationResult) {
	if _, exist := e.traceResults[node]; exist {
		delete(e.traceResults, node)
	}
	value := formatValueFromResult(e.masker, result)
	e.traceResults[node] = value
}

func (e *EvaluationContext) tryGetTraceResult(node IExpressionNode) (exist bool, result string) {
	result, exist = e.traceResults[node]
	return
}
