package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"

	"github.com/dungdm93/drasi/pkg/expr"
)

type result struct {
	kind        expr.ResultKind
	value       any
	level       int
	omitTracing bool
}

func newEvaluationResultWithTrace(eCtx interfaces.Context, level int, val any, kind expr.ResultKind) *result {
	return newEvaluationResult(eCtx, level, val, kind, false)
}

func newEvaluationResultSkipTrace(eCtx interfaces.Context, level int, val any, kind expr.ResultKind) *result {
	return newEvaluationResult(eCtx, level, val, kind, true)
}

func newEvaluationResult(eCtx interfaces.Context, level int, val any, kind expr.ResultKind, omitTracing bool) *result {
	e := &result{
		kind:        kind,
		value:       val,
		level:       level,
		omitTracing: omitTracing,
	}
	if !omitTracing {
		e.traceValue(eCtx)
	}
	return e
}

func (e *result) Value() any {
	return e.value
}

func (e *result) Kind() expr.ResultKind {
	return e.kind
}

func (e *result) traceValue(eCtx interfaces.Context) {
	if !e.omitTracing && eCtx.Masker() != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", common.FormatValue(eCtx.Masker(), e.value, e.kind)))
	}
}

func (e *result) traceVerbose(eCtx interfaces.Context, msg string) {
	padding := strings.Repeat(".", e.level*2) // Use strings.Repeat for padding
	if !e.omitTracing {
		eCtx.Trace().Debug(fmt.Sprintf("%s%s", padding, msg))
	}
}

func (e *result) isFalsy() bool {
	switch e.kind {
	case expr.Null:
		return true
	case expr.Boolean:
		return !e.value.(bool)
	case expr.Number:
		numb := e.value.(float64)
		return numb == 0 || math.IsNaN(numb)
	case expr.String:
		str := e.value.(string)
		return strings.EqualFold(str, "")
	default:
		return false
	}
}

func (e *result) isTruthy() bool {
	return !e.isFalsy()
}

func (e *result) equal(right expr.Result) bool {
	return equal(e.value, right.Value())
}

func equal(canonicalLeftValue, canonicalRightValue any) bool {
	canonicalLeftValue, canonicalRightValue, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case expr.Null:
			// Null, Null
			return true
		case expr.Number:
			// Number, Number
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l == r
		case expr.String:
			// String, String
			lStr := canonicalLeftValue.(string)
			rStr := canonicalRightValue.(string)
			return strings.EqualFold(lStr, rStr)
		case expr.Boolean:
			// Boolean, Boolean
			lB := canonicalLeftValue.(bool)
			rB := canonicalRightValue.(bool)
			return lB == rB
		case expr.Object, expr.Array:
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

func coerceTypes(canonicalLeftValue, canonicalRightValue any) (leftValue, rightValue any, lk, rk expr.ResultKind) {
	lk = getKind(canonicalLeftValue)
	rk = getKind(canonicalRightValue)
	if lk == rk {
		// Same kind
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Number, String
	if lk == expr.Number && rk == expr.String {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		rk = expr.Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// String, Number
	if lk == expr.String && rk == expr.Number {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		lk = expr.Number
		return canonicalLeftValue, canonicalRightValue, lk, rk
	}
	// Boolean|Null, Any
	if lk == expr.Boolean || lk == expr.Null {
		canonicalLeftValue = convertToNumber(canonicalLeftValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	// Any, Boolean|Null
	if rk == expr.Boolean || rk == expr.Null {
		canonicalRightValue = convertToNumber(canonicalRightValue)
		return coerceTypes(canonicalLeftValue, canonicalRightValue)
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func convertToNumber(canonicalValue any) float64 {
	kind := getKind(canonicalValue)
	switch kind {
	case expr.Null:
		return 0
	case expr.Boolean:
		if canonicalValue.(bool) {
			return 1
		}
		return 0
	case expr.Number:
		return canonicalValue.(float64)
	case expr.String:
		return common.ParseNumber(canonicalValue.(string))
	}
	return math.NaN()
}

func getKind(canonicalValue any) expr.ResultKind {
	if canonicalValue == nil {
		return expr.Null
	}
	if _, isBool := canonicalValue.(bool); isBool {
		return expr.Boolean
	}
	if _, isFloat64 := canonicalValue.(float64); isFloat64 {
		return expr.Number
	}
	if _, isStr := canonicalValue.(string); isStr {
		return expr.String
	}
	if _, isObj := canonicalValue.(common.Obj); isObj {
		return expr.Object
	}
	if _, isArr := canonicalValue.(common.Array); isArr {
		return expr.Array
	}
	return expr.Object
}

func (e *result) greaterThan(right expr.Result) bool {
	_, _, lk, rk := coerceTypes(e.value, right.Value())
	if lk == rk {
		switch lk {
		case expr.Number:
			// Number & Number
			l := e.value.(float64)
			r := right.Value().(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l > r
		case expr.String:
			// String & String
			return e.value.(string) > right.Value().(string)
		case expr.Boolean:
			// Boolean & Boolean
			return e.value.(bool) && !right.Value().(bool)
		}
	}
	return false
}

func (e *result) greaterThanOrEqual(right expr.Result) bool {
	return e.equal(right) || e.greaterThan(right)
}

func (e *result) lessThan(right expr.Result) bool {
	return lessThan(e.value, right.Value())
}

func lessThan(canonicalLeftValue, canonicalRightValue any) bool {
	_, _, lk, rk := coerceTypes(canonicalLeftValue, canonicalRightValue)
	if lk == rk {
		switch lk {
		case expr.Number:
			l := canonicalLeftValue.(float64)
			r := canonicalRightValue.(float64)
			if math.IsNaN(l) || math.IsNaN(r) {
				return false
			}
			return l < r
		case expr.String:
			return canonicalLeftValue.(string) < canonicalRightValue.(string)
		case expr.Boolean:
			return !canonicalLeftValue.(bool) && canonicalRightValue.(bool)
		}
	}
	return false
}

func (e *result) lessThanOrEqual(right expr.Result) bool {
	return e.equal(right) || e.lessThan(right)
}

func (e *result) notEqual(right expr.Result) bool {
	return !e.equal(right)
}

func (e *result) number() float64 {
	return convertToNumber(e.value)
}

func (e *result) string() string {
	switch e.kind {
	case expr.Null:
		return ""
	case expr.Boolean:
		if e.value.(bool) {
			return common.True
		}
		return common.False
	case expr.Number:
		d := e.value.(float64)
		return fmt.Sprintf("%f", d)
	case expr.String:
		return e.value.(string)
	default:
		return e.kind.ToString()
	}
}

func (e *result) primitive() bool {
	return isPrimitive(e.kind)
}

// createIntermediateResult is useful for working with values that are not the direct evaluation result of a parameter.
// This allows Node authors to leverage the coercion and comparison functions
// for any values.
//
// Also note, the value will be canonicalized (for example numeric types converted to double) and any
// matching interfaces applied.
func createIntermediateResult(eCtx interfaces.Context, obj any) *result {
	val, kind := common.ToCanonicalValue(obj)
	return newEvaluationResultSkipTrace(eCtx, 0, val, kind)
}

// isCollection perform type assert if result's value implement Obj or Array and return corresponding interface
func (e *result) isCollection() (ok bool, collection any) {
	if e.kind == expr.Object || e.kind == expr.Array {
		obj := e.value
		o, isObj := obj.(common.Obj)
		if isObj {
			return true, o
		}
		a, isArr := obj.(common.Array)
		if isArr {
			return true, a
		}
	}
	return false, nil
}
