package workflows

import (
	"github.com/dungdm93/drasi/pkg/model"
	"reflect"
)

var (
	typeDecoder          = reflect.TypeFor[model.Decoder]()
	typeEvaluableBool    = reflect.TypeFor[Evaluable[bool]]()
	typeEvaluableInteger = reflect.TypeFor[Evaluable[int64]]()
	typeEvaluableFloat   = reflect.TypeFor[Evaluable[float64]]()
	typeEvaluableString  = reflect.TypeFor[Evaluable[string]]()
	typeConditional      = reflect.TypeFor[Conditional]()
)

func DecodeEvaluableHook(fromType reflect.Type, toType reflect.Type, data any) (any, error) {
	if toType.Implements(typeDecoder) {
		// let's toType decode its self
		return data, nil
	}

	if toType.Implements(typeEvaluableBool) {
		if fromType.Kind() == reflect.Bool {
			b := data.(bool)
			return NewIdent(b), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return NewEvaluable(s, toBool)
		}
	} else if toType.Implements(typeEvaluableInteger) {
		if fromType.Kind() >= reflect.Int && fromType.Kind() <= reflect.Uint64 {
			i := data.(int64)
			return NewIdent(i), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return NewEvaluable(s, toInteger)
		}
	} else if toType.Implements(typeEvaluableFloat) {
		if fromType.Kind() == reflect.Float32 || fromType.Kind() == reflect.Float64 {
			f := data.(float64)
			return NewIdent(f), nil
		}
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return NewEvaluable(s, toFloat)
		}
	} else if toType.Implements(typeEvaluableString) {
		if fromType.Kind() == reflect.String {
			s := data.(string)
			return NewEvaluable(s, toString)
		}
	}

	return data, nil
}

func DecodeConditionalHook(fromType reflect.Type, toType reflect.Type, data any) (any, error) {
	if toType.Implements(typeDecoder) {
		// let's toType decode its self
		return data, nil
	}
	if !toType.Implements(typeConditional) || fromType.Kind() != reflect.String {
		return data, nil
	}

	s := data.(string)
	return NewConditional(s), nil
}

func init() {
	model.RegisterDecodeHook(DecodeEvaluableHook)
	model.RegisterDecodeHook(DecodeConditionalHook)
}
