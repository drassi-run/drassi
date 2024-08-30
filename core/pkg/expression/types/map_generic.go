package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func NewMapGeneric[M ~map[K]V, K comparable, V any](value M) ref.Val {
	indexType := genericToType[K]()
	if indexType == ref.TypeInvalid {
		return NewError("%w: map key %s", errUnsupportedType, reflect.TypeFor[K]())
	}

	return &Map{
		mapAccessor: &genericMapAccessor[M, K, V]{
			indexType: indexType,
			instance:  value,
		},
		value: value,
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

func (a *genericMapAccessor[M, K, V]) Size() int {
	return len(a.instance)
}

func (a *genericMapAccessor[M, K, V]) IndexType() ref.Type {
	return a.indexType
}

func (a *genericMapAccessor[M, K, V]) Get(index any) ref.Val {
	idx, ok := index.(K)
	if !ok {
		return NewError("%w: index must be %s, got %T", errInvalidType, reflect.TypeFor[K](), index)
	}
	return a.get(idx)
}

func (a *genericMapAccessor[M, K, V]) get(idx K) ref.Val {
	v, ok := a.instance[idx]
	if !ok {
		return NULL
	}
	return NativeToVal(v)
}

func (a *genericMapAccessor[M, K, V]) Items() traits.Iterator {
	return func(yield func(ref.Val, ref.Val) bool) {
		for k, v := range a.instance {
			key, value := NativeToVal(k), NativeToVal(v)
			if !yield(key, value) {
				return
			}
		}
	}
}
