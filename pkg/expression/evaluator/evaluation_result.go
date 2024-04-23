package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/shared"
)

type EvaluationResult struct {
	kind        shared.ValueKind
	raw         any
	value       any
	level       int
	omitTracing bool
}

func NewEvaluationResultWithTrace(eCtx interfaces.IEvaluationContext, level int, val any, kind shared.ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, false)
}

func NewEvaluationResultSkipTrace(eCtx interfaces.IEvaluationContext, level int, val any, kind shared.ValueKind, raw any) *EvaluationResult {
	return NewEvaluationResult(eCtx, level, val, kind, raw, true)
}

func NewEvaluationResult(eCtx interfaces.IEvaluationContext, level int, val any, kind shared.ValueKind, raw any, omitTracing bool) *EvaluationResult {
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

func (e *EvaluationResult) GetKind() shared.ValueKind {
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

func (e *EvaluationResult) traceValue1(eCtx interfaces.IEvaluationContext, val any, kind shared.ValueKind) {
	if !e.omitTracing && eCtx.Masker() != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", formatValue(eCtx.Masker(), val, kind)))
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
	case shared.ValueKindNull:
		return true
	case shared.ValueKindBoolean:
		return !e.value.(bool)
	case shared.ValueKindNumber:
		numb := e.value.(float64)
		return numb == 0 || math.IsNaN(numb)
	case shared.ValueKindString:
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
		case shared.ValueKindNull:
			// ValueKindNull, ValueKindNull
			return true
		case shared.ValueKindNumber:
			// ValueKindNumber, ValueKindNumber
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l == r
		case shared.ValueKindString:
			// ValueKindString, ValueKindString
			lStr := canonicalLeftValue.(string)
			rStr := canonicalRightValue.(string)
			return strings.EqualFold(lStr, rStr)
		case shared.ValueKindBoolean:
			// ValueKindBoolean, ValueKindBoolean
			lB := canonicalLeftValue.(bool)
			rB := canonicalRightValue.(bool)
			return lB == rB
		case shared.ValueKindObject, shared.ValueKindArray:
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

func coerceTypes(canonicalLeftValue, canonicalRightValue any) (leftValue, rightValue any, lk, rk shared.ValueKind) {
	lk = getKind(canonicalLeftValue)
	rk = getKind(canonicalRightValue)
	if lk == rk {
		// Same kind
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindNumber, ValueKindString
	if lk == shared.ValueKindNumber && rk == shared.ValueKindString {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		rk = shared.ValueKindNumber
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindString, ValueKindNumber
	if lk == shared.ValueKindString && rk == shared.ValueKindNumber {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		lk = shared.ValueKindNumber
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// ValueKindBoolean|ValueKindNull, Any
	if lk == shared.ValueKindBoolean || lk == shared.ValueKindNull {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	// Any, ValueKindBoolean|ValueKindNull
	if rk == shared.ValueKindBoolean || rk == shared.ValueKindNull {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func convertToNumber(canonicalValue any) float64 {
	kind := getKind(canonicalValue)
	switch kind {
	case shared.ValueKindNull:
		return 0
	case shared.ValueKindBoolean:
		if canonicalValue.(bool) {
			return 1
		}
		return 0
	case shared.ValueKindNumber:
		return canonicalValue.(float64)
	case shared.ValueKindString:
		return parseNumber(canonicalValue.(string))
	}
	return math.NaN()
}

func getKind(canonicalValue any) shared.ValueKind {
	if canonicalValue == nil {
		return shared.ValueKindNull
	}
	if _, isBool := canonicalValue.(bool); isBool {
		return shared.ValueKindBoolean
	}
	if _, isFloat64 := canonicalValue.(float64); isFloat64 {
		return shared.ValueKindNumber
	}
	if _, isStr := canonicalValue.(string); isStr {
		return shared.ValueKindString
	}
	if _, isObj := canonicalValue.(interfaces.IReadOnlyObj); isObj {
		return shared.ValueKindObject
	}
	if _, isArr := canonicalValue.(interfaces.IReadOnlyArray); isArr {
		return shared.ValueKindArray
	}
	return shared.ValueKindObject
}

func (e *EvaluationResult) AbstractGreaterThan(right interfaces.IEvaluationResult) bool {
	_, _, lk, rk := coerceTypes(e.value, right.Value())
	if lk == rk {
		switch lk {
		case shared.ValueKindNumber:
			// ValueKindNumber & ValueKindNumber
			l := e.value.(float64)
			r := right.Value().(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l > r
		case shared.ValueKindString:
			// ValueKindString & ValueKindString
			return e.value.(string) > right.Value().(string)
		case shared.ValueKindBoolean:
			// ValueKindBoolean & ValueKindBoolean
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
		case shared.ValueKindNumber:
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l < r
		case shared.ValueKindString:
			return canonicalLeftValue.(string) < canonicalRightValue.(string)
		case shared.ValueKindBoolean:
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
	case shared.ValueKindNull:
		return ""
	case shared.ValueKindBoolean:
		if e.value.(bool) {
			return shared.True
		}
		return shared.False
	case shared.ValueKindNumber:
		d := e.value.(float64)
		return fmt.Sprintf("%f", d)
	case shared.ValueKindString:
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
func createIntermediateResult(eCtx interfaces.IEvaluationContext, obj any) *EvaluationResult {
	val, kind, raw := convertToCanonicalValue(obj)
	return NewEvaluationResultSkipTrace(eCtx, 0, val, kind, raw)
}

// TryGetCollectionInterface perform type assert if EvaluationResult's value implement IReadOnlyObj or IReadOnlyArray
// and
// return
// corresponding interface
func (e *EvaluationResult) TryGetCollectionInterface() (ok bool, collection any) {
	if e.kind == shared.ValueKindObject || e.kind == shared.ValueKindArray {
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
