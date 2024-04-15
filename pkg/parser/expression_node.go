package parser

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dungdm93/drasi/pkg/parser/constants"
)

type (
	IExpressionNode interface {
		convertToExpression() string
		convertToRealizedExpression(eCtx *EvaluationContext) string
		// createMemoryCounter(eCtx *EvaluationContext) *MemoryCounter
		evaluateCore(eCtx *EvaluationContext) any
		setContainer(c iContainer)
		getContainer() iContainer
		getName() string
		setName(name string)
		traceFullyRealized() bool
		getLevel() (level int)
	}
	ExpressionNode struct {
		container iContainer
		IExpressionNode
		level int
		// name is used for tracing. Normally the parser will set the name. However, if a node
		// is added manually, then the name may not be set and will fall back to the type name.
		name string
	}
)

func Evaluate(e IExpressionNode, trace ITraceWriter, masker ISecretMasker, state any, opt *EvaluationOption) *EvaluationResult {
	if e.getContainer() != nil {
		panic(fmt.Errorf("evaluate can only be called from root node"))
	}
	if masker != nil {
		masker = masker.Clone()
	} else {
		masker = NewSecretMasker()
	}
	eTrace := NewEvaluationTraceWriter(trace, masker)
	eCtx := NewEvaluationContext(eTrace, masker, state, opt, e)
	eCtx.trace.Info(fmt.Sprintf("Evaluating: %s", e.convertToExpression()))
	result := evaluate(eCtx, e)
	traceTreeResult(eCtx, e, result.value, result.kind)
	return result
}

func traceTreeResult(eCtx *EvaluationContext, e IExpressionNode, result any, kind ValueKind) {
	realizedExp := e.convertToRealizedExpression(eCtx)
	traceValue := formatValue(eCtx.masker, result, kind)
	if !strings.EqualFold(realizedExp, traceValue) {
		if kind == Number && realizedExp == fmt.Sprintf("'%s'", traceValue) {
			// Don't bother tracing the realized expression when the result is a number and the
			// realized expression is a precisely matching string.
		} else {
			eCtx.trace.Info(fmt.Sprintf("Expanded: %s", realizedExp))
		}
	}
	eCtx.trace.Info(fmt.Sprintf("Result: %s", traceValue))
}

func evaluate(eCtx *EvaluationContext, e IExpressionNode) *EvaluationResult {
	var level int
	if e.getContainer() != nil {
		level = e.getContainer().getLevel() + 1
	}
	coreResult := e.evaluateCore(eCtx)

	_, kind, raw := convertToCanonicalValue(coreResult)
	result := NewEvaluationResultSkipTrace(eCtx, level, coreResult, kind, raw)
	if e.traceFullyRealized() {
		eCtx.setTraceResult(e, result)
	}
	return result
}

func convertToCanonicalValue(input any) (value any, kind ValueKind, raw any) {
	if input == nil {
		kind = Null
		return
	}
	if _, castable := input.(bool); castable {
		kind = Boolean
		value = input
		return
	}
	if _, castable := input.(float64); castable {
		value = input
		kind = Number
		return
	}
	if _, castable := input.(string); castable {
		kind = String
		value = input
		return
	}
	if b, castable := input.(IBool); castable {
		kind = Boolean
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(INumber); castable {
		kind = Number
		raw = input
		value = b.GetValue()
		return
	}
	if b, castable := input.(IString); castable {
		kind = String
		raw = input
		value = b.GetValue()
		return
	}
	if _, castable := input.(IReadOnlyObj); castable {
		kind = Object
		value = input
		return
	}
	if _, castable := input.(IReadOnlyArray); castable {
		kind = Array
		value = input
		return
	}
	if _, castable := input.(INull); castable {
		kind = Null
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
			kind = Number
			value, err := strconv.ParseFloat(input.(string), 64)
			if err != nil {
				panic(err)
			}
			return value, kind, nil
		}
	}
	kind = Object
	value = input
	return
}

func newNodeFromToken(tk *LexicalToken) IExpressionNode {
	switch tk.Kind() {
	case LTKStartIndex, LTKDereference:
		return new(Index)
	case LTKLogicalOperator:
		switch tk.RawValue() {
		case constants.Not:
			return new(Not)
		case constants.NotEqual:
			return new(NotEqual)
		case constants.GreaterThan:
			return new(GreaterThan)
		case constants.GreaterThanOrEqual:
			return new(GreaterThanOrEqual)
		case constants.LessThan:
			return new(LessThan)
		case constants.LessThanOrEqual:
			return new(LessThanOrEqual)
		case constants.Equal:
			return new(Equal)
		case constants.And:
			return new(And)
		case constants.Or:
			return new(Or)
		default:
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", tk.RawValue()))
		}
	case LTKNull, LTKBoolean, LTKNumber, LTKString:
		return NewLiteral(tk.ParsedValue())
	case LTKPropertyName:
		return NewLiteral(tk.RawValue())
	case LTKWildcard:
		return new(WildCard)
	}
	panic(fmt.Errorf("unexpected kind %s when creating node", tk.Kind()))
}
