package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"golang.org/x/exp/maps"
)

func NewMapGeneric[M ~map[K]V, K comparable, V any](value M) ref.Val {
	indexType := genericToType[K]()
	if indexType == ref.TypeInvalid {
		return NewError("unsupported map key %s: %w", reflect.TypeFor[K](), errUnsupportedType)
	}

	return &Map{
		mapAccessor: &genericMapAccessor[M, K, V]{
			indexType: indexType,
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

func (a *genericMapAccessor[M, K, V]) Get(index any) ref.Val {
	idx, ok := index.(K)
	if !ok {
		return WrapError(errInvalidType)
	}
	return a.get(idx)
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
