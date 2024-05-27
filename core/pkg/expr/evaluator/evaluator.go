package evaluator

import (
	"fmt"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/common"
	"github.com/dungdm93/drassi/core/pkg/secret_masker"
)

func EvaluateWithDefaults(root ast_ifaces.ExprNode, state any, workingDir string) (r expr.Result, err error) {
	return Evaluate(root, nil, state, &Option{WorkingDir: workingDir})
}

// Evaluate from ast root and return an error if any
func Evaluate(root ast_ifaces.ExprNode, masker secret_masker.Interface, state any, opt *Option) (r expr.Result, err error) {
	defer func() {
		if v := recover(); v != nil {
			r, err = nil, fmt.Errorf("%+v", v)
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
	eCtx := newContext(eTrace, masker, state, opt, &visitorImpl{workingDir: opt.WorkingDir})
	eCtx.traceWriter.Info(fmt.Sprintf("Evaluating: %s", root.Expr()))
	result := evaluate(eCtx, root)
	traceTreeResult(eCtx, root, result.value, result.kind)
	return result, nil
}

// evaluate traverse the ast and call ast_ifaces.Accept() on every node
func evaluate(eCtx ast_ifaces.Context, node ast_ifaces.ExprNode) *result {
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

// traceTreeResult helps to understand what expression is expanded
func traceTreeResult(eCtx *context, node ast_ifaces.ExprNode, result any, kind expr.ResultKind) {
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
