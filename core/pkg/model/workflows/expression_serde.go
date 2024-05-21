package workflows

import (
	"reflect"
	"strings"

	"github.com/dungdm93/drassi/core/pkg/model"
)

func (e *Evaluable[R]) DecodeMapstructure(a any) (any, error) {
	return map[string]any{"token": a}, nil
}

var typeToken = reflect.TypeFor[Token]()

func DecodeTokenHook(from reflect.Value, to reflect.Value) (any, error) {
	if !to.Type().Implements(typeToken) || to.Interface() != nil {
		return valueOf(from), nil
	}

	var (
		token Token = nil
		data  any   = nil
	)
	switch from.Kind() {
	case reflect.Bool:
		token = NewLiteralToken(from.Bool())
	case reflect.String:
		s := from.String()
		if strings.Contains(s, OpenExpression) {
			token = NewExpressionToken(s)
		} else {
			token = NewLiteralToken(s)
		}
	case reflect.Slice, reflect.Array:
		token = NewSequenceToken(nil)
		data = from.Interface()
	case reflect.Map:
		token = NewMappingToken(nil)
		data = from.Interface()
	default:
		if from.CanInt() {
			token = NewLiteralToken(from.Int())
		} else if from.CanUint() {
			token = NewLiteralToken(from.Uint())
		} else if from.CanFloat() {
			token = NewLiteralToken(from.Float())
		} else {
			data = valueOf(from)
		}
	}
	if token != nil {
		to.Set(reflect.ValueOf(token))
	}
	return data, nil
}

func (m *mappingToken) DecodeMapstructure(input any) (any, error) {
	inputVal := reflect.ValueOf(input)
	if inputVal.Kind() != reflect.Map {
		return input, nil
	}

	a := make([]KVPair[any, any], 0, inputVal.Len())
	mapIter := inputVal.MapRange()
	for mapIter.Next() {
		key := mapIter.Key()
		val := mapIter.Value()
		pair := KVPair[any, any]{
			Key:   key.Interface(),
			Value: val.Interface(),
		}
		a = append(a, pair)
	}
	return a, nil
}

func init() {
	model.RegisterDecodeHook(DecodeTokenHook)
}
