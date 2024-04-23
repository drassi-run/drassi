package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression"
)

type EvaluationResult struct {
	kind        expression.ValueKind
	raw         any
	value       any
	level       int
	omitTracing bool
}

func newEvaluationResultWithTrace(eCtx expression.IEvaluationContext, level int, val any, kind expression.ValueKind, raw any) *EvaluationResult {
	return newEvaluationResult(eCtx, level, val, kind, raw, false)
}

func newEvaluationResultSkipTrace(eCtx expression.IEvaluationContext, level int, val any, kind expression.ValueKind, raw any) *EvaluationResult {
	return newEvaluationResult(eCtx, level, val, kind, raw, true)
}

func newEvaluationResult(eCtx expression.IEvaluationContext, level int, val any, kind expression.ValueKind, raw any, omitTracing bool) *EvaluationResult {
	e := &EvaluationResult{
		kind:        kind,
		raw:         raw,
		value:       val,
		level:       level,
		omitTracing: omitTracing,
	}
	if !omitTracing {
		e.traceValue(eCtx)
	}
	return e
}

func (e *EvaluationResult) Raw() any {
	return e.raw
}

func (e *EvaluationResult) Value() any {
	return e.value
}

func (e *EvaluationResult) GetKind() expression.ValueKind {
	return e.kind
}

func (e *EvaluationResult) Level() int {
	return e.level
}

func (e *EvaluationResult) traceValue(eCtx expression.IEvaluationContext) {
	if !e.omitTracing && eCtx.Masker() != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", formatValue(eCtx.Masker(), e.value, e.kind)))
	}

}

func (e *EvaluationResult) traceVerbose(eCtx expression.IEvaluationContext, msg string) {
	padding := strings.Repeat(".", e.level*2) // Use strings.Repeat for padding
	if !e.omitTracing {
		eCtx.Trace().Verbose(fmt.Sprintf("%s%s", padding, msg))
	}
}

func (e *EvaluationResult) IsFalsy() bool {
	switch e.kind {
	case expression.ValueKindNull:
		return true
	case expression.ValueKindBoolean:
		return !e.value.(bool)
	case expression.ValueKindNumber:
		numb := e.value.(float64)
		return numb == 0 || math.IsNaN(numb)
	case expression.ValueKindString:
		str := e.value.(string)
		return strings.EqualFold(str, "")
	default:
		return false
	}
}

func (e *EvaluationResult) IsTruthy() bool {
	return !e.IsFalsy()
}

func (e *EvaluationResult) AbstractEqual(right expression.IEvaluationResult) bool {
	return abstractEqual(e.value, right.Value())
}

func abstractEqual(canonicalLeftValue, canonicalRightValue any) bool {
	canonicalLeftValue, canonicalRightValue, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case expression.ValueKindNull:
			// ValueKindNull, ValueKindNull
			return true
		case expression.ValueKindNumber:
			// ValueKindNumber, ValueKindNumber
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l == r
		case expression.ValueKindString:
			// ValueKindString, ValueKindString
			lStr := canonicalLeftValue.(string)
			rStr := canonicalRightValue.(string)
			return strings.EqualFold(lStr, rStr)
		case expression.ValueKindBoolean:
			// ValueKindBoolean, ValueKindBoolean
			lB := canonicalLeftValue.(bool)
			rB := canonicalRightValue.(bool)
			return lB == rB
		case expression.ValueKindObject, expression.ValueKindArray:
			if reflect.ValueOf(canonicalLeftValue).IsZero() || reflect.ValueOf(canonicalRightValue).IsZero() {
				// zero value of them same kind are considered equal
				return true
			}
			lGK := reflect.TypeOf(canonicalLeftValue).Kind()
			switch lGK {
			case reflect.Array:
				return reflect.DeepEqual(canonicalLeftValue, canonicalRightValue)
			case reflect.Map, reflect.Struct, reflect.Slice, reflect.Pointer:
				return reflect.ValueOf(canonicalLeftValue).Pointer() == reflect.ValueOf(canonicalRightValue).Pointer()
			}
		}
	}
	return false
}

func coerceTypes(canonicalLeftValue, canonicalRightValue any) (leftValue, rightValue any, lk, rk expression.ValueKind) {
	lk = getKind(canonicalLeftValue)
	rk = getKind(canonicalRightValue)
	if lk == rk {
		// Same kind
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindNumber, ValueKindString
	if lk == expression.ValueKindNumber && rk == expression.ValueKindString {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		rk = expression.ValueKindNumber
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindString, ValueKindNumber
	if lk == expression.ValueKindString && rk == expression.ValueKindNumber {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		lk = expression.ValueKindNumber
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindBoolean|ValueKindNull, Any
	if lk == expression.ValueKindBoolean || lk == expression.ValueKindNull {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	// Any, ValueKindBoolean|ValueKindNull
	if rk == expression.ValueKindBoolean || rk == expression.ValueKindNull {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func convertToNumber(canonicalValue any) float64 {
	kind := getKind(canonicalValue)
	switch kind {
	case expression.ValueKindNull:
		return 0
	case expression.ValueKindBoolean:
		if canonicalValue.(bool) {
			return 1
		}
		return 0
	case expression.ValueKindNumber:
		return canonicalValue.(float64)
	case expression.ValueKindString:
		return parseNumber(canonicalValue.(string))
	}
	return math.NaN()
}

func getKind(canonicalValue any) expression.ValueKind {
	if canonicalValue == nil {
		return expression.ValueKindNull
	}
	if _, isBool := canonicalValue.(bool); isBool {
		return expression.ValueKindBoolean
	}
	if _, isFloat64 := canonicalValue.(float64); isFloat64 {
		return expression.ValueKindNumber
	}
	if _, isStr := canonicalValue.(string); isStr {
		return expression.ValueKindString
	}
	if _, isObj := canonicalValue.(expression.IReadOnlyObj); isObj {
		return expression.ValueKindObject
	}
	if _, isArr := canonicalValue.(expression.IReadOnlyArray); isArr {
		return expression.ValueKindArray
	}
	return expression.ValueKindObject
}

func (e *EvaluationResult) AbstractGreaterThan(right expression.IEvaluationResult) bool {
	_, _, lk, rk := coerceTypes(e.value, right.Value())
	if lk == rk {
		switch lk {
		case expression.ValueKindNumber:
			// ValueKindNumber & ValueKindNumber
			l := e.value.(float64)
			r := right.Value().(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l > r
		case expression.ValueKindString:
			// ValueKindString & ValueKindString
			return e.value.(string) > right.Value().(string)
		case expression.ValueKindBoolean:
			// ValueKindBoolean & ValueKindBoolean
			return e.value.(bool) && !right.Value().(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractGreaterThanOrEqual(right expression.IEvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractGreaterThan(right)
}

func (e *EvaluationResult) AbstractLessThan(right expression.IEvaluationResult) bool {
	return abstractLessThan(e.value, right.Value())
}

func abstractLessThan(canonicalLeftValue, canonicalRightValue any) bool {
	_, _, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case expression.ValueKindNumber:
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l < r
		case expression.ValueKindString:
			return canonicalLeftValue.(string) < canonicalRightValue.(string)
		case expression.ValueKindBoolean:
			return !canonicalLeftValue.(bool) && canonicalRightValue.(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractLessThanOrEqual(right expression.IEvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractLessThan(right)
}

func (e *EvaluationResult) AbstractNotEqual(right expression.IEvaluationResult) bool {
	return !e.AbstractEqual(right)
}

func (e *EvaluationResult) ConvertToNumber() float64 {
	return convertToNumber(e.value)
}

func (e *EvaluationResult) ConvertToString() string {
	switch e.kind {
	case expression.ValueKindNull:
		return ""
	case expression.ValueKindBoolean:
		if e.value.(bool) {
			return expression.True
		}
		return expression.False
	case expression.ValueKindNumber:
		d := e.value.(float64)
		return fmt.Sprintf("%f", d)
	case expression.ValueKindString:
		return e.value.(string)
	default:
		return e.kind.ToString()
	}
}

func (e *EvaluationResult) IsPrimitive() bool {
	return isPrimitive(e.kind)
}

// createIntermediateResult is useful for working with values that are not the direct evaluation result of a parameter.
// This allows ExpressionNodeBs authors to leverage the coercion and comparison functions
// for any values.
//
// Also note, the value will be canonicalized (for example numeric types converted to double) and any
// matching interfaces applied.
func createIntermediateResult(eCtx expression.IEvaluationContext, obj any) *EvaluationResult {
	val, kind, raw := convertToCanonicalValue(obj)
	return newEvaluationResultSkipTrace(eCtx, 0, val, kind, raw)
}

// TryGetCollectionInterface perform type assert if EvaluationResult's value implement IReadOnlyObj or IReadOnlyArray and return corresponding interface
func (e *EvaluationResult) TryGetCollectionInterface() (ok bool, collection any) {
	if e.kind == expression.ValueKindObject || e.kind == expression.ValueKindArray {
		obj := e.value
		o, isObj := obj.(expression.IReadOnlyObj)
		if isObj {
			return true, o
		}
		a, isArr := obj.(expression.IReadOnlyArray)
		if isArr {
			return true, a
		}
	}
	return false, nil
}
