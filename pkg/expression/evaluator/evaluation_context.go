package evaluator

import (
	"errors"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

type (
	evaluationContext struct {
		state        any
		opt          *EvaluationOption
		traceWriter  expression.ITraceWriter
		secretMasker secret_masker.ISecretMasker
		traceResults map[expression.IExpNode]string
	}
)

var (
	ErrorsEmptyTrace        = errors.New("traceWriter must be provided")
	ErrorsEmptySecretMasker = errors.New("secret secretMasker must be provider")
)

func newEvaluationContext(trace expression.ITraceWriter, masker secret_masker.ISecretMasker, state any, opt *EvaluationOption,
	node expression.IExpNode) *evaluationContext {
	if trace == nil {
		panic(ErrorsEmptyTrace)
	}
	if masker == nil {
		panic(ErrorsEmptySecretMasker)
	}
	e := &evaluationContext{
		state:        state,
		traceWriter:  trace,
		secretMasker: masker,
	}
	e.opt = opt
	e.traceResults = map[expression.IExpNode]string{}
	return e
}

func (e *evaluationContext) State() any {
	return e.state
}

func (e *evaluationContext) Masker() secret_masker.ISecretMasker {
	return e.secretMasker
}

func (e *evaluationContext) Trace() expression.ITraceWriter {
	return e.traceWriter
}

func (e *evaluationContext) SetTraceResult(node expression.IExpNode, result expression.IEvaluationResult) {
	if _, exist := e.traceResults[node]; exist {
		delete(e.traceResults, node)
	}
	value := formatValueFromResult(e.secretMasker, result)
	e.traceResults[node] = value
}

func (e *evaluationContext) TryGetTraceResult(node expression.IExpNode) (exist bool, result string) {
	result, exist = e.traceResults[node]
	return
}
