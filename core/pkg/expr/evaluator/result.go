package evaluator

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/expr"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/common"
)

type result struct {
	kind        expr.ResultKind
	value       any
	level       int
	omitTracing bool
}

func newEvaluationResultWithTrace(eCtx ast_ifaces.Context, level int, val any, kind expr.ResultKind) *result {
	return newEvaluationResult(eCtx, level, val, kind, false)
}

func newEvaluationResultSkipTrace(eCtx ast_ifaces.Context, level int, val any, kind expr.ResultKind) *result {
	return newEvaluationResult(eCtx, level, val, kind, true)
}

func newEvaluationResult(eCtx ast_ifaces.Context, level int, val any, kind expr.ResultKind, omitTracing bool) *result {
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
	var (
		pmaxInt32 = 2147483647.0
		nmaxInt32 = -2147483648.0
	)

	if e.kind == expr.Number && reflect.TypeOf(e.value).Name() == reflect.TypeFor[float64]().Name() {
		f := e.value.(float64)
		if f < nmaxInt32 || f > pmaxInt32 {
			return e.value
		}
		floored := int(f)
		if float64(floored) == f {
			return floored
		}
	}
	return e.value
}

func (e *result) Kind() expr.ResultKind {
	return e.kind
}

func (e *result) traceValue(eCtx ast_ifaces.Context) {
	if !e.omitTracing && eCtx.Masker() != nil {
		e.traceVerbose(eCtx, fmt.Sprintf("=> %s", common.FormatValue(eCtx.Masker(), e.value, e.kind)))
	}
}

func (e *result) traceVerbose(eCtx ast_ifaces.Context, msg string) {
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
		switch e.value.(type) {
		case float64:
			numb := e.value.(float64)
			return numb == 0 || math.IsNaN(numb)
		case int:
			numb := e.value.(int)
			return numb == 0
		default:
			panic("isFalsy should not reach here")
		}
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
			switch canonicalLeftValue.(type) {
			// canonicalLeftValue & canonicalRightValue should be of same kind and type
			case float64:
				// Number, Number
				l := canonicalLeftValue.(float64)
				r := canonicalRightValue.(float64)
				if math.IsNaN(l) || math.IsNaN(r) {
					return false
				}
				return l == r
			case int:
				// Number, Number
				l := canonicalLeftValue.(int)
				r := canonicalRightValue.(int)
				return l == r
			}
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
		// Same kind different type
		if lk == expr.Number && reflect.TypeOf(canonicalLeftValue).Name() != reflect.TypeOf(canonicalRightValue).Name() {
			lf, lFloatAble := canonicalLeftValue.(float64)
			if lFloatAble {
				// left float, right int
				// floor left
				intLeft := int(lf)
				// floored left == left
				if float64(intLeft) == canonicalLeftValue.(float64) {
					return intLeft, canonicalRightValue, lk, rk
				} else {
					return lf, float64(canonicalRightValue.(int)), lk, rk
				}
			}
			rf, rFloatAble := canonicalRightValue.(float64)
			if rFloatAble {
				// left int, right float
				// floor right
				intRight := int(rf)
				// floored right == right
				if float64(intRight) == canonicalRightValue.(float64) {
					return canonicalLeftValue, intRight, lk, rk
				} else {
					return float64(canonicalLeftValue.(int)), rf, lk, rk
				}
			}
		} else {
			// Same kind same type
			return canonicalLeftValue, canonicalRightValue, lk, rk
		}
	}
	// Number, String
	if lk == expr.Number && rk == expr.String {
		// coerce based on type of left operand
		switch canonicalLeftValue.(type) {
		case int:
			canonicalRightValue = toInt(canonicalRightValue)
			rk = expr.Number
			return canonicalLeftValue, canonicalRightValue, lk, rk
		case float64:
			canonicalRightValue = toFloat(canonicalRightValue)
			rk = expr.Number
			return canonicalLeftValue, canonicalRightValue, lk, rk
		}
	}
	// String, Number
	if lk == expr.String && rk == expr.Number {
		// coerce based on type of right operand
		switch canonicalRightValue.(type) {
		case int:
			canonicalLeftValue = toInt(canonicalLeftValue)
			lk = expr.Number
			return canonicalLeftValue, canonicalRightValue, lk, rk
		case float64:
			canonicalLeftValue = toFloat(canonicalLeftValue)
			lk = expr.Number
			return canonicalLeftValue, canonicalRightValue, lk, rk
		}
	}
	// Boolean|Null, Any
	if lk == expr.Boolean || lk == expr.Null {
		// coerce based on type of right operand
		switch canonicalRightValue.(type) {
		case int:
			canonicalLeftValue = toInt(canonicalLeftValue)
			return coerceTypes(canonicalLeftValue, canonicalRightValue)
		case float64:
			canonicalLeftValue = toFloat(canonicalLeftValue)
			return coerceTypes(canonicalLeftValue, canonicalRightValue)
		case string:
			if canonicalRightValue.(string) == "" {
				return coerceTypes(canonicalLeftValue, nil)
			}
		}
	}
	// Any, Boolean|Null
	if rk == expr.Boolean || rk == expr.Null {
		switch canonicalLeftValue.(type) {
		case int:
			canonicalRightValue = toInt(canonicalRightValue)
			return coerceTypes(canonicalLeftValue, canonicalRightValue)
		case float64:
			canonicalRightValue = toFloat(canonicalRightValue)
			return coerceTypes(canonicalLeftValue, canonicalRightValue)
		case string:
			if canonicalLeftValue.(string) == "" {
				return coerceTypes(nil, canonicalRightValue)
			}
		}
	}
	return canonicalLeftValue, canonicalRightValue, lk, rk
}

func toFloat(canonicalValue any) float64 {
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
		return common.ParseFloat(canonicalValue.(string))
	}
	return math.NaN()
}

func toInt(canonicalValue any) int {
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
		return canonicalValue.(int)
	case expr.String:
		i, err := common.ParseInt(canonicalValue.(string))
		if err == nil {
			return i
		}
	}
	panic("toInt should not reach here")
}

func getKind(canonicalValue any) expr.ResultKind {
	switch canonicalValue.(type) {
	case nil:
		return expr.Null
	case bool:
		return expr.Boolean
	case float64, int:
		return expr.Number
	case string:
		return expr.String
	case common.Array:
		return expr.Array
	default:
		return expr.Object
	}
	// if canonicalValue == nil {
	// 	return expr.Null
	// }
	// if _, isBool := canonicalValue.(bool); isBool {
	// 	return expr.Boolean
	// }
	// if _, isFloat64 := canonicalValue.(float64); isFloat64 {
	// 	return expr.Number
	// }
	// if _, isStr := canonicalValue.(string); isStr {
	// 	return expr.String
	// }
	// if _, isObj := canonicalValue.(common.Obj); isObj {
	// 	return expr.Object
	// }
	// if _, isArr := canonicalValue.(common.Array); isArr {
	// 	return expr.Array
	// }
	// return expr.Object
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
	return toFloat(e.value)
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
		switch e.value.(type) {
		case float64:
			f := e.value.(float64)
			// chooses between fixed-point and exponential notation based on the value's magnitude and precision, aiming to present the number in the most compact form without losing significant digits.
			return fmt.Sprintf("%g", f)
		case int:
			i := e.value.(int)
			return fmt.Sprintf("%d", i)
		default:
			panic("result.string() should not reach here")
		}
	case expr.String:
		return e.value.(string)
	default:
		return e.kind.String()
	}
}

func (e *result) primitive() bool {
	return isPrimitive(e.kind)
}

// createIntermediateResult is useful for working with values that are not the direct evaluation result of a parameter.
// This allows ExprNode authors to leverage the coercion and comparison functions
// for any values.
//
// Also note, the value will be canonicalized (for example numeric types converted to double) and any
// matching ast_ifaces applied.
func createIntermediateResult(eCtx ast_ifaces.Context, obj any) *result {
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

func isPrimitive(kind expr.ResultKind) bool {
	switch kind {
	case expr.Null, expr.Boolean, expr.Number, expr.String:
		return true
	default:
		return false
	}
}
