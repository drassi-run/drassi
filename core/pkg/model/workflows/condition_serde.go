package workflows

import (
	"reflect"

	"github.com/dungdm93/drassi/core/pkg/model"
)

var typeConditional = reflect.TypeFor[Conditional]()

func DecodeConditionalHook(fromType reflect.Type, toType reflect.Type, data any) (any, error) {
	if !toType.Implements(typeConditional) || fromType.Kind() != reflect.String {
		return data, nil
	}

	s := data.(string)
	return NewConditional(s), nil
}

func init() {
	model.RegisterDecodeHook(DecodeConditionalHook)
}
