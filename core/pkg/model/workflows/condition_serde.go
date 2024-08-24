package workflows

import (
	"reflect"

	"drassi.run/core/pkg/model"
)

var typeConditional = reflect.TypeFor[Conditional]()

func DecodeConditionalHook(fromType reflect.Type, toType reflect.Type, data any) (any, error) {
	if !toType.Implements(typeConditional) {
		return data, nil
	}

	if s, ok := model.Stringify(data); ok {
		return NewConditional(s), nil
	}
	return data, nil
}

func init() {
	model.RegisterDecodeHook(DecodeConditionalHook)
}
