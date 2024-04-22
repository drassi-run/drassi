package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

type EvaluationResult struct {
	kind        interfaces.ValueKind
	raw         any
	value       any
	level       int
	omitTracing bool
}

func NewEvaluationResultWithTrace(eCtx interfaces.IEvaluationContext, level int, val any, kind interfaces.ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, false)
}

func NewEvaluationResultSkipTrace(eCtx interfaces.IEvaluationContext, level int, val any, kind interfaces.ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, true)
}

func NewEvaluationResult(eCtx interfaces.IEvaluationContext, level int, val any, kind interfaces.ValueKind, raw any, omitTracing bool) *EvaluationResult {
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

func (e *EvaluationResult) GetKind() interfaces.ValueKind {
	return e.kind
}

func (e *EvaluationResult) Level() int {
	return e.level
}

func (e *EvaluationResult) traceValue(eCtx interfaces.IEvaluationContext) {
	if !e.omitTracing {
		e.traceValue1(eCtx, e.value, e.kind)
	}

}

func (e *EvaluationResult) traceValue1(eCtx interfaces.IEvaluationContext, val any, kind interfaces.ValueKind) {
	if !e.omitTracing && eCtx.Masker() != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", FormatValue(eCtx.Masker(), val, kind)))
	}
}

func (e *EvaluationResult) traceVerbose(eCtx interfaces.IEvaluationContext, msg string) {
	padding := strings.Repeat(".", e.level*2) // Use strings.Repeat for padding
	if !e.omitTracing {
		eCtx.Trace().Verbose(fmt.Sprintf("%s%s", padding, msg))
	}
}

func (e *EvaluationResult) IsFalsy() bool {
	switch e.kind {
	case interfaces.Null:
		return true
	case interfaces.Boolean:
		return !e.value.(bool)
	case interfaces.Number:
		numb := e.value.(float64)
		return numb == 0 || math.IsNaN(numb)
	case interfaces.String:
		str := e.value.(string)
		return strings.EqualFold(str, "")
	default:
		return false
	}
}

func (e *EvaluationResult) IsTruthy() bool {
	return !e.IsFalsy()
}

func (e *EvaluationResult) AbstractEqual(right interfaces.IEvaluationResult) bool {
	return abstractEqual(e.value, right.Value())
}

func abstractEqual(canonicalLeftValue, canonicalRightValue any) bool {
	canonicalLeftValue, canonicalRightValue, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case interfaces.Null:
			// Null, Null
			return true
		case interfaces.Number:
			// Number, Number
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l == r
		case interfaces.String:
			// String, String
			lStr := canonicalLeftValue.(string)
			rStr := canonicalRightValue.(string)
			return strings.EqualFold(lStr, rStr)
		case interfaces.Boolean:
			// Boolean, Boolean
			lB := canonicalLeftValue.(bool)
			rB := canonicalRightValue.(bool)
			return lB == rB
		case interfaces.Object, interfaces.Array:
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

func coerceTypes(canonicalLeftValue, canonicalRightValue any) (leftValue, rightValue any, lk, rk interfaces.ValueKind) {
	lk = getKind(canonicalLeftValue)
	rk = getKind(canonicalRightValue)
	if lk == rk {
		// Same kind
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Number, String
	if lk == interfaces.Number && rk == interfaces.String {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		rk = interfaces.Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// String, Number
	if lk == interfaces.String && rk == interfaces.Number {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		lk = interfaces.Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Boolean|Null, Any
	if lk == interfaces.Boolean || lk == interfaces.Null {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	// Any, Boolean|Null
	if rk == interfaces.Boolean || rk == interfaces.Null {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func convertToNumber(canonicalValue any) float64 {
	kind := getKind(canonicalValue)
	switch kind {
	case interfaces.Null:
		return 0
	case interfaces.Boolean:
		if canonicalValue.(bool) {
			return 1
		}
		return 0
	case interfaces.Number:
		return canonicalValue.(float64)
	case interfaces.String:
		return ParseNumber(canonicalValue.(string))
	}
	return math.NaN()
}

func getKind(canonicalValue any) interfaces.ValueKind {
	if canonicalValue == nil {
		return interfaces.Null
	}
	if _, isBool := canonicalValue.(bool); isBool {
		return interfaces.Boolean
	}
	if _, isFloat64 := canonicalValue.(float64); isFloat64 {
		return interfaces.Number
	}
	if _, isStr := canonicalValue.(string); isStr {
		return interfaces.String
	}
	if _, isObj := canonicalValue.(interfaces.IReadOnlyObj); isObj {
		return interfaces.Object
	}
	if _, isArr := canonicalValue.(interfaces.IReadOnlyArray); isArr {
		return interfaces.Array
	}
	return interfaces.Object
}

func (e *EvaluationResult) AbstractGreaterThan(right interfaces.IEvaluationResult) bool {
	_, _, lk, rk := coerceTypes(e.value, right.Value())
	if lk == rk {
		switch lk {
		case interfaces.Number:
			// Number & Number
			l := e.value.(float64)
			r := right.Value().(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l > r
		case interfaces.String:
			// String & String
			return e.value.(string) > right.Value().(string)
		case interfaces.Boolean:
			// Boolean & Boolean
			return e.value.(bool) && !right.Value().(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractGreaterThanOrEqual(right interfaces.IEvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractGreaterThan(right)
}

func (e *EvaluationResult) AbstractLessThan(right interfaces.IEvaluationResult) bool {
	return abstractLessThan(e.value, right.Value())
}

func abstractLessThan(canonicalLeftValue, canonicalRightValue any) bool {
	_, _, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case interfaces.Number:
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l < r
		case interfaces.String:
			return canonicalLeftValue.(string) < canonicalRightValue.(string)
		case interfaces.Boolean:
			return !canonicalLeftValue.(bool) && canonicalRightValue.(bool)
		}
	}
	return false
}

func (e *EvaluationResult) AbstractLessThanOrEqual(right interfaces.IEvaluationResult) bool {
	return e.AbstractEqual(right) || e.AbstractLessThan(right)
}

func (e *EvaluationResult) AbstractNotEqual(right interfaces.IEvaluationResult) bool {
	return !e.AbstractEqual(right)
}

func (e *EvaluationResult) ConvertToNumber() float64 {
	return convertToNumber(e.value)
}

func (e *EvaluationResult) ConvertToString() string {
	switch e.kind {
	case interfaces.Null:
		return ""
	case interfaces.Boolean:
		if e.value.(bool) {
			return constants.True
		}
		return constants.False
	case interfaces.Number:
		d := e.value.(float64)
		return fmt.Sprintf("%f", d)
	case interfaces.String:
		return e.value.(string)
	default:
		return e.kind.ToString()
	}
}

func (e *EvaluationResult) IsPrimitive() bool {
	return IsPrimitive(e.kind)
}

// CreateIntermediateResult is useful for working with values that are not the direct evaluation result of a parameter.
// This allows ExpressionNodeBase authors to leverage the coercion and comparison functions
// for any values.
//
// Also note, the value will be canonicalized (for example numeric types converted to double) and any
// matching interfaces applied.
func CreateIntermediateResult(eCtx interfaces.IEvaluationContext, obj any) *EvaluationResult {
	val, kind, raw := ConvertToCanonicalValue(obj)
	return NewEvaluationResultSkipTrace(eCtx, 0, val, kind, raw)
}

// TryGetCollectionInterface perform type assert if EvaluationResult's value implement IReadOnlyObj or IReadOnlyArray
// and
// return
// corresponding interface
func (e *EvaluationResult) TryGetCollectionInterface() (ok bool, collection any) {
	if e.kind == interfaces.Object || e.kind == interfaces.Array {
		obj := e.value
		o, isObj := obj.(interfaces.IReadOnlyObj)
		if isObj {
			return true, o
		}
		a, isArr := obj.(interfaces.IReadOnlyArray)
		if isArr {
			return true, a
		}
	}
	return false, nil
}
