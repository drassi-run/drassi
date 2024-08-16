package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func NewMapDynamic(value any) ref.Val {
	instance := reflect.ValueOf(value)

	if instance.Kind() != reflect.Map {
		return NewError("expect a map, got %T: %w", value, errInvalidType)
	}

	keyType := instance.Type().Key()
	indexType := reflectToType(keyType)

	if indexType == ref.TypeInvalid {
		return NewError("unsupported map key %s: %w", keyType, errUnsupportedType)
	}

	return &Map{
		mapAccessor: &dynamicMapAccessor{
			indexType: indexType,
			keyType:   keyType,
			instance:  instance,
		},
		value: value,
	}
}

func reflectToType(t reflect.Type) ref.Type {
	switch t.Kind() {
	case reflect.String:
		return ref.TypeString
	case reflect.Bool:
		return ref.TypeBoolean
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ref.TypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ref.TypeInteger
	case reflect.Float32, reflect.Float64:
		return ref.TypeFloat
	case reflect.Struct:
		return ref.TypeStruct
	case reflect.Map:
		return ref.TypeMap
	case reflect.Array, reflect.Slice:
		return ref.TypeList
	default:
		return ref.TypeInvalid
	}
}

type dynamicMapAccessor struct {
	indexType ref.Type
	keyType   reflect.Type
	instance  reflect.Value
}

func (a *dynamicMapAccessor) Size() int {
	return a.instance.Len()
}

func (a *dynamicMapAccessor) IndexType() ref.Type {
	return a.indexType
}

func (a *dynamicMapAccessor) Get(index any) ref.Val {
	idx := reflect.ValueOf(index)
	if idx.Type() != a.keyType {
		if !idx.Type().ConvertibleTo(a.keyType) {
			return WrapError(errInvalidType)
		}
		idx = idx.Convert(a.keyType)
	}

	return a.get(idx)
}

func (a *dynamicMapAccessor) get(idx reflect.Value) ref.Val {
	v := a.instance.MapIndex(idx)
	return NativeToVal(v)
}

func (a *dynamicMapAccessor) Iterator() traits.Iterator {
	return func(yield func(ref.Val, ref.Val) bool) {
		it := a.instance.MapRange()
		for it.Next() {
			k, v := it.Key(), it.Value()
			key, value := NativeToVal(k), NativeToVal(v)
			if !yield(key, value) {
				return
			}
		}
	}
}
