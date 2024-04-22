package evaluator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser"
)

func EvaluateWithContext(eCtx interfaces.IEvaluationContext, e interfaces.IExpressionNode) *EvaluationResult {
	// visitor := new(ExpressionNodeVisitor)

	var level int
	if e.GetContainer() != nil {
		level = e.GetContainer().GetLevel() + 1
	}

	coreResult := e.EvaluateCore(eCtx)

	_, kind, raw := ConvertToCanonicalValue(coreResult)
	result := NewEvaluationResultSkipTrace(eCtx, level, coreResult, kind, raw)
	if e.TraceFullyRealized() {
		eCtx.SetTraceResult(e, result)
	}
	return result
}

func Evaluate(e interfaces.IExpressionNode, trace interfaces.ITraceWriter, masker interfaces.ISecretMasker, state any, opt *EvaluationOption) *EvaluationResult {
	if e.GetContainer() != nil {
		panic(errors.New("evaluate can only be called from root node"))
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = parser.NewSecretMasker()
	}
	eTrace := parser.NewEvaluationTraceWriter(trace, masker)
	eCtx := NewEvaluationContext(eTrace, masker, state, opt, e)
	eCtx.trace.Info(fmt.Sprintf("Evaluating: %s", e.ConvertToExpression()))
	result := EvaluateWithContext(eCtx, e)
	TraceTreeResult(eCtx, e, result.value, result.kind)
	return result
}

func TraceTreeResult(eCtx *EvaluationContext, e interfaces.IExpressionNode, result any, kind interfaces.ValueKind) {
	realizedExp := e.ConvertToRealizedExpression(eCtx)
	traceValue := FormatValue(eCtx.masker, result, kind)
	if !strings.EqualFold(realizedExp, traceValue) {
		if kind == interfaces.Number && realizedExp == fmt.Sprintf("'%s'", traceValue) {
			// Don't bother tracing the realized expression when the result is a number and the
			// realized expression is a precisely matching string.
		} else {
			eCtx.trace.Info(fmt.Sprintf("Expanded: %s", realizedExp))
		}
	}
	eCtx.trace.Info(fmt.Sprintf("Result: %s", traceValue))
}

func ConvertToCanonicalValue(input any) (value any, kind interfaces.ValueKind, raw any) {
	if input == nil {
		kind = interfaces.Null
		return
	}
	if _, castable := input.(bool); castable {
		kind = interfaces.Boolean
		value = input
		return
	}
	if _, castable := input.(float64); castable {
		value = input
		kind = interfaces.Number
		return
	}
	if _, castable := input.(string); castable {
		kind = interfaces.String
		value = input
		return
	}
	if b, castable := input.(interfaces.IBool); castable {
		kind = interfaces.Boolean
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(interfaces.INumber); castable {
		kind = interfaces.Number
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(interfaces.IString); castable {
		kind = interfaces.String
		raw = input
		value = b.GetValue()
		return
	}
	if _, castable := input.(interfaces.IReadOnlyObj); castable {
		kind = interfaces.Object
		value = input
		return
	}
	if _, castable := input.(interfaces.IReadOnlyArray); castable {
		kind = interfaces.Array
		value = input
		return
	}
	if _, castable := input.(interfaces.INull); castable {
		kind = interfaces.Null
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
			kind = interfaces.Number
			value, err := strconv.ParseFloat(input.(string), 64)
			if err != nil {
				panic(err)
			}
			return value, kind, nil
		}
	}
	kind = interfaces.Object
	value = input
	return
}
