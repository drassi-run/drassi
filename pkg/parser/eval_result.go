package parser

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/parser/constants"
)

type EvaluationResult struct {
	kind        ValueKind
	raw         any
	value       any
	level       int
	omitTracing bool
}

func NewEvaluationResultWithTrace(eCtx *EvaluationContext, level int, val any, kind ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, false)
}

func NewEvaluationResultSkipTrace(eCtx *EvaluationContext, level int, val any, kind ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, true)
}

func NewEvaluationResult(eCtx *EvaluationContext, level int, val any, kind ValueKind, raw any, omitTracing bool) *EvaluationResult {
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

func (e *EvaluationResult) Level() int {
	return e.level
}

func (e *EvaluationResult) traceValue(eCtx *EvaluationContext) {
	if !e.omitTracing {
		e.traceValue1(eCtx, e.value, e.kind)
	}

}

func (e *EvaluationResult) traceValue1(eCtx *EvaluationContext, val any, kind ValueKind) {
	if !e.omitTracing && eCtx.masker != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", formatValue(eCtx.masker, val, kind)))
	}
}

func (e *EvaluationResult) traceVerbose(eCtx *EvaluationContext, msg string) {
	padding := strings.Repeat(".", e.level*2) // Use strings.Repeat for padding
	if !e.omitTracing {
		eCtx.trace.Verbose(fmt.Sprintf("%s%s", padding, msg))
	}
}

func (e *EvaluationResult) IsFalsy() bool {
	switch e.kind {
	case Null:
		return true
	case Boolean:
		return !e.value.(bool)
	case Number:
		numb := e.value.(float64)
		return numb == 0 || math.IsNaN(numb)
	case String:
		str := e.value.(string)
		return strings.EqualFold(str, "")
	default:
		return false
	}
}

func (e *EvaluationResult) IsTruthy() bool {
	return !e.IsFalsy()
}

func (e *EvaluationResult) AbstractEqual(right *EvaluationResult) bool {
	return abstractEqual(e.value, right.value)
}

func abstractEqual(canonicalLeftValue, canonicalRightValue any) bool {
	canonicalLeftValue, canonicalRightValue, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case Null:
			// Null, Null
			return true
		case Number:
			// Number, Number
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l == r
		case String:
			// String, String
			lStr := canonicalLeftValue.(string)
			rStr := canonicalRightValue.(string)
			return strings.EqualFold(lStr, rStr)
		case Boolean:
			// Boolean, Boolean
			lB := canonicalLeftValue.(bool)
			rB := canonicalRightValue.(bool)
			return lB == rB
		case Object, Array:
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

func coerceTypes(canonicalLeftValue, canonicalRightValue any) (leftValue, rightValue any, lk, rk ValueKind) {
	lk = getKind(canonicalLeftValue)
	rk = getKind(canonicalRightValue)
	if lk == rk {
		// Same kind
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Number, String
	if lk == Number && rk == String {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		rk = Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// String, Number
	if lk == String && rk == Number {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		lk = Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Boolean|Null, Any
	if lk == Boolean || lk == Null {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	// Any, Boolean|Null
	if rk == Boolean || rk == Null {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func convertToNumber(canonicalValue any) float64 {
	kind := getKind(canonicalValue)
	switch kind {
	case Null:
		return 0
	case Boolean:
		if canonicalValue.(bool) {
			return 1
		}
		return 0
	case Number:
		return canonicalValue.(float64)
	case String:
		return ParseNumber(canonicalValue.(string))
	}
	return math.NaN()
}

func getKind(canonicalValue any) ValueKind {
	if canonicalValue == nil {
		return Null
	}
	if _, isBool := canonicalValue.(bool); isBool {
		return Boolean
	}
	if _, isFloat64 := canonicalValue.(float64); isFloat64 {
		return Number
	}
	if _, isStr := canonicalValue.(string); isStr {
		return String
	}
	if _, isObj := canonicalValue.(IReadOnlyObj); isObj {
		return Object
	}
	if _, isArr := canonicalValue.(IReadOnlyArray); isArr {
		return Array
	}
	return Object
}

func (e *EvaluationResult) AbstractGreaterThan(right *EvaluationResult) bool {
	_, _, lk, rk := coerceTypes(e.value, right.value)
	if lk == rk {
		switch lk {
		case Number:
			// Number & Number
			l := e.value.(float64)
			r := right.value.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l > r
		case String:
			// String & String
			return e.value.(string) > right.value.(string)
		case Boolean:
			// Boolean & Boolean
			return e.value.(bool) && !right.value.(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractGreaterThanOrEqual(right *EvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractGreaterThan(right)
}

func (e *EvaluationResult) AbstractLessThan(right *EvaluationResult) bool {
	return abstractLessThan(e.value, right.value)
}

func abstractLessThan(canonicalLeftValue, canonicalRightValue any) bool {
	_, _, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case Number:
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l < r
		case String:
			return canonicalLeftValue.(string) < canonicalRightValue.(string)
		case Boolean:
			return !canonicalLeftValue.(bool) && canonicalRightValue.(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractLessThanOrEqual(right *EvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractLessThan(right)
}

func (e *EvaluationResult) AbstractNotEqual(right *EvaluationResult) bool {
	return !e.AbstractEqual(right)
}

func (e *EvaluationResult) ConvertToNumber() float64 {
	return convertToNumber(e.value)
}

func (e *EvaluationResult) ConvertToString() string {
	switch e.kind {
	case Null:
		return ""
	case Boolean:
		if e.value.(bool) {
			return constants.True
		}
		return constants.False
	case Number:
		d := e.value.(float64)
		return fmt.Sprintf("%f", d)
	case String:
		return e.value.(string)
	default:
		return e.kind.ToString()
	}
}

func (e *EvaluationResult) IsPrimitive() bool {
	return isPrimitive(e.kind)
}

// CreateIntermediateResult is useful for working with values that are not the direct evaluation result of a parameter.
// This allows ExpressionNode authors to leverage the coercion and comparison functions
// for any values.
//
// Also note, the value will be canonicalized (for example numeric types converted to double) and any
// matching interfaces applied.
func CreateIntermediateResult(eCtx *EvaluationContext, obj any) *EvaluationResult {
	val, kind, raw := convertToCanonicalValue(obj)
	return NewEvaluationResultSkipTrace(eCtx, 0, val, kind, raw)
}

// TryGetCollectionInterface perform type assert if EvaluationResult's value implement IReadOnlyObj or IReadOnlyArray
// and
// return
// corresponding interface
func (e *EvaluationResult) TryGetCollectionInterface() (ok bool, collection any) {
	if e.kind == Object || e.kind == Array {
		obj := e.value
		o, isObj := obj.(IReadOnlyObj)
		if isObj {
			return true, o
		}
		a, isArr := obj.(IReadOnlyArray)
		if isArr {
			return true, a
		}
	}
	return false, nil
}
