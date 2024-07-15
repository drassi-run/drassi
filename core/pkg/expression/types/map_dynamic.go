package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"fmt"
	"reflect"
)

func NewMapDynamic(value any) *Map {
	instance := reflect.ValueOf(value)
	keyType := instance.Type().Key()
	return &Map{
		mapAccessor: &dynamicMapAccessor{
			indexType: reflectToType(keyType),
			keyType:   keyType,
			instance:  instance,
		},
		value: value,
		size:  instance.Len(),
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
		return -1
	}
}

type dynamicMapAccessor struct {
	indexType ref.Type
	keyType   reflect.Type
	instance  reflect.Value
}

func (a *dynamicMapAccessor) IndexType() ref.Type {
	return a.indexType
}

func (a *dynamicMapAccessor) Get(index any) (ref.Val, error) {
	idx := reflect.ValueOf(index)
	if !idx.Type().ConvertibleTo(a.keyType) {
		return nil, fmt.Errorf("invalid index type")
	}

	idx = idx.Convert(a.keyType)
	return a.get(idx), nil
}

func (a *dynamicMapAccessor) get(idx reflect.Value) ref.Val {
	value := a.instance.MapIndex(idx)
	v := value.Interface()
	return NativeToVal(v)
}

func (a *dynamicMapAccessor) Iterator() traits.Iterator {
	return &mapIterator[reflect.Value]{
		getter: a.get,
		keys:   a.instance.MapKeys(),
		cursor: 0,
	}
}
