package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"golang.org/x/exp/maps"
)

func NewMapGeneric[M ~map[K]V, K comparable, V any](value M) *Map {
	return &Map{
		mapAccessor: &genericMapAccessor[M, K, V]{
			indexType: genericToType[K](),
			instance:  value,
		},
		value: value,
		size:  len(value),
	}
}

func genericToType[K comparable]() ref.Type {
	t := reflect.TypeFor[K]()
	return reflectToType(t)
}

type genericMapAccessor[M ~map[K]V, K comparable, V any] struct {
	indexType ref.Type
	instance  M
}

func (a *genericMapAccessor[M, K, V]) IndexType() ref.Type {
	return a.indexType
}

func (a *genericMapAccessor[M, K, V]) Get(index any) (ref.Val, error) {
	idx, ok := index.(K)
	if !ok {
		return nil, errInvalidType
	}
	return a.get(idx), nil
}

func (a *genericMapAccessor[M, K, V]) get(idx K) ref.Val {
	e := a.instance[idx]
	return NativeToVal(e)
}

func (a *genericMapAccessor[M, K, V]) Iterator() traits.Iterator {
	return &mapIterator[K]{
		getter: a.get,
		keys:   maps.Keys(a.instance),
		cursor: 0,
	}
}
