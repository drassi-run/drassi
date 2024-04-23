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

func evaluateWithContext(eCtx expression.IEvaluationContext, e expression.IExpressionNode) *EvaluationResult {
	visitor := new(expressionNodeVisitor)

	var level int
	if e.GetContainer() != nil {
		level = e.GetContainer().GetLevel() + 1
	}

	coreResult := e.Accept(eCtx, visitor)
	_, kind, raw := convertToCanonicalValue(coreResult)
	result := NewEvaluationResultSkipTrace(eCtx, level, coreResult, kind, raw)
	if e.TraceFullyRealized() {
		eCtx.SetTraceResult(e, result)
	}
	return result
}

func Evaluate(e expression.IExpressionNode, trace expression.ITraceWriter, masker interfaces.ISecretMasker, state any, opt *EvaluationOption) *EvaluationResult {
	if e.GetContainer() != nil {
		panic(errors.New("evaluate can only be called from root node"))
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = newNoOpSecretMasker()
	}
	eTrace := newEvaluationTraceWriter(trace, masker)
	eCtx := newEvaluationContext(eTrace, masker, state, opt, e)
	eCtx.trace.Info(fmt.Sprintf("Evaluating: %s", e.ConvertToExpression()))
	result := evaluateWithContext(eCtx, e)
	traceTreeResult(eCtx, e, result.value, result.kind)
	return result
}

func traceTreeResult(eCtx *evaluationContext, e expression.IExpressionNode, result any, kind expression.ValueKind) {
	realizedExp := e.ConvertToRealizedExpression(eCtx)
	traceValue := formatValue(eCtx.masker, result, kind)
	if !strings.EqualFold(realizedExp, traceValue) {
		if kind == expression.ValueKindNumber && realizedExp == fmt.Sprintf("'%s'", traceValue) {
			// Don't bother tracing the realized expression when the result is a number and the
			// realized expression is a precisely matching string.
		} else {
			eCtx.trace.Info(fmt.Sprintf("Expanded: %s", realizedExp))
		}
	}
	eCtx.trace.Info(fmt.Sprintf("Result: %s", traceValue))
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
	if b, castable := input.(expression.IBool); castable {
		kind = expression.ValueKindBoolean
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(expression.INumber); castable {
		kind = expression.ValueKindNumber
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(expression.IString); castable {
		kind = expression.ValueKindString
		raw = input
		value = b.GetValue()
		return
	}
	if _, castable := input.(expression.IReadOnlyObj); castable {
		kind = expression.ValueKindObject
		value = input
		return
	}
	if _, castable := input.(expression.IReadOnlyArray); castable {
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
