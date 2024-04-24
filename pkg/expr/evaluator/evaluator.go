package evaluator

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

// MustEvaluate from ast root and panic if there is an error
func MustEvaluate(root interfaces.Node, masker secret_masker.SecretMasker, state any,
	opt *Option) expr.Result {
	if root.GetCtn() != nil {
		panic(ErrorNonRootEvaluate)
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = newNoOpSecretMasker()
	}
	eTrace := newEvaluationTraceWriter(masker)
	eCtx := newContext(eTrace, masker, state, opt, new(evaluationVisitor))
	eCtx.traceWriter.Info(fmt.Sprintf("Evaluating: %s", root.Expr()))
	result := evaluate(eCtx, root)
	traceTreeResult(eCtx, root, result.value, result.kind)
	return result
}

// Evaluate from ast root and return an error if any
func Evaluate(root interfaces.Node, masker secret_masker.SecretMasker, state any,
	opt *Option) (r expr.Result, err error) {
	defer func() {
		if v := recover(); v != nil {
			r, err = nil, fmt.Errorf("%v", v)
		}
	}()
	if root.GetCtn() != nil {
		panic(ErrorNonRootEvaluate)
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = newNoOpSecretMasker()
	}
	eTrace := newEvaluationTraceWriter(masker)
	eCtx := newContext(eTrace, masker, state, opt, new(evaluationVisitor))
	eCtx.traceWriter.Info(fmt.Sprintf("Evaluating: %s", root.Expr()))
	result := evaluate(eCtx, root)
	traceTreeResult(eCtx, root, result.value, result.kind)
	return result, nil
}

// evaluate traverse the ast and call interfaces.Accept() on every node
func evaluate(eCtx interfaces.Context, node interfaces.Node) *result {
	var level int
	if node.GetCtn() != nil {
		level = node.GetCtn().GetLevel() + 1
	}
	coreResult := node.Accept(eCtx, eCtx.Visitor())
	_, kind := common.ToCanonicalValue(coreResult)
	result := newEvaluationResultWithTrace(eCtx, level, coreResult, kind)
	if node.TraceFullyRealized() {
		eCtx.SetTraceResult(node, result)
	}
	return result
}

func traceTreeResult(eCtx *context, node interfaces.Node, result any, kind expr.ResultKind) {
	realizedExp := node.RealizedExpr(eCtx)
	traceValue := common.FormatValue(eCtx.secretMasker, result, kind)
	if !strings.EqualFold(realizedExp, traceValue) {
		if kind == expr.Number && realizedExp == fmt.Sprintf("'%s'", traceValue) {
			// Don't bother tracing the realized expr when the result is a number and the
			// realized expr is a precisely matching string.
		} else {
			eCtx.traceWriter.Info(fmt.Sprintf("Expanded: %s", realizedExp))
		}
	}
	eCtx.traceWriter.Info(fmt.Sprintf("Result: %s", traceValue))
}
