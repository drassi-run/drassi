package evaluator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/secret_masker"
)

func Evaluate(node expression.IExpNode, trace expression.ITraceWriter, masker secret_masker.ISecretMasker, state any, opt *EvaluationOption) *EvaluationResult {
	if node.GetContainer() != nil {
		panic(errors.New("evaluate can only be called from root node"))
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = newNoOpSecretMasker()
	}
	eTrace := newEvaluationTraceWriter(trace, masker)
	eCtx := newEvaluationContext(eTrace, masker, state, opt, node)
	eCtx.traceWriter.Info(fmt.Sprintf("Evaluating: %s", node.ConvertToExpression()))
	result := evaluateWithContext(eCtx, node)
	traceTreeResult(eCtx, node, result.value, result.kind)
	return result
}

func evaluateWithContext(eCtx expression.IEvaluationContext, node expression.IExpNode) *EvaluationResult {
	visitor := new(expNodeVisitor)
	var level int
	if node.GetContainer() != nil {
		level = node.GetContainer().GetLevel() + 1
	}
	coreResult := node.Accept(eCtx, visitor)
	_, kind, raw := convertToCanonicalValue(coreResult)
	result := newEvaluationResultWithTrace(eCtx, level, coreResult, kind, raw)
	if node.TraceFullyRealized() {
		eCtx.SetTraceResult(node, result)
	}
	return result
}

func traceTreeResult(eCtx *evaluationContext, node expression.IExpNode, result any, kind expression.ValueKind) {
	realizedExp := node.ConvertToRealizedExpression(eCtx)
	traceValue := formatValue(eCtx.secretMasker, result, kind)
	if !strings.EqualFold(realizedExp, traceValue) {
		if kind == expression.ValueKindNumber && realizedExp == fmt.Sprintf("'%s'", traceValue) {
			// Don't bother tracing the realized expression when the result is a number and the
			// realized expression is a precisely matching string.
		} else {
			eCtx.traceWriter.Info(fmt.Sprintf("Expanded: %s", realizedExp))
		}
	}
	eCtx.traceWriter.Info(fmt.Sprintf("Result: %s", traceValue))
}

func convertToCanonicalValue(input any) (value any, kind expression.ValueKind, raw any) {
	if input == nil {
		kind = expression.ValueKindNull
		return
	}
	if _, castable := input.(bool); castable {
		kind = expression.ValueKindBoolean
		value = input
		return
	}
	if _, castable := input.(float64); castable {
		value = input
		kind = expression.ValueKindNumber
		return
	}
	if _, castable := input.(string); castable {
		kind = expression.ValueKindString
		value = input
		return
	}
	if _, castable := input.(expression.ReadOnlyObj); castable {
		kind = expression.ValueKindObject
		value = input
		return
	}
	if _, castable := input.(expression.ReadOnlyArray); castable {
		kind = expression.ValueKindArray
		value = input
		return
	}
	if _, castable := input.(expression.INull); castable {
		kind = expression.ValueKindNull
		raw = input
		value = nil
		return
	}
	if !(reflect.TypeOf(input).Kind() == reflect.Struct) {
		_, isInt := input.(int)
		_, isInt8 := input.(int8)
		_, isInt16 := input.(int16)
		_, isInt32 := input.(int32)
		_, isInt64 := input.(int64)

		_, isUint := input.(uint)
		_, isUint8 := input.(uint8)
		_, isUint16 := input.(uint16)
		_, isUint32 := input.(uint32)
		_, isUint64 := input.(uint64)

		_, isFloat32 := input.(float32)
		var isWellKnownNumber bool
		if isInt || isInt8 || isInt16 || isInt32 || isInt64 {
			isWellKnownNumber = true
		}
		if isUint || isUint8 || isUint16 || isUint32 || isUint64 {
			isWellKnownNumber = true
		}
		if isFloat32 {
			isWellKnownNumber = true
		}
		if isWellKnownNumber {
			kind = expression.ValueKindNumber
			value, err := strconv.ParseFloat(input.(string), 64)
			if err != nil {
				panic(err)
			}
			return value, kind, nil
		}
	}
	kind = expression.ValueKindObject
	value = input
	return
}
