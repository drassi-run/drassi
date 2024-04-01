package model

import "reflect"

var (
	typeEvaluableBool    = reflect.TypeFor[Evaluable[bool]]()
	typeEvaluableInteger = reflect.TypeFor[Evaluable[int64]]()
	typeEvaluableFloat   = reflect.TypeFor[Evaluable[float64]]()
	typeEvaluableString  = reflect.TypeFor[Evaluable[string]]()
)

func DecodeEvaluableHook(fromType reflect.Type, toType reflect.Type, data any) (any, error) {
	if toType.Implements(typeEvaluableBool) {
		if fromType.Kind() == reflect.Bool {
			b := data.(bool)
			return newIdent(b), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return newEvaluable(s, toBool)
		}
	} else if toType.Implements(typeEvaluableInteger) {
		if fromType.Kind() >= reflect.Int && fromType.Kind() <= reflect.Uint64 {
			i := data.(int64)
			return newIdent(i), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return newEvaluable(s, toInteger)
		}
	} else if toType.Implements(typeEvaluableFloat) {
		if fromType.Kind() == reflect.Float32 || fromType.Kind() == reflect.Float64 {
			f := data.(float64)
			return newIdent(f), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return newEvaluable(s, toFloat)
		}
	} else if toType.Implements(typeEvaluableString) {
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return newEvaluable(s, toString)
		}
	}

	return data, nil
}
