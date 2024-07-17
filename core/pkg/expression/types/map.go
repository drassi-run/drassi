package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

type mapAccessor interface {
	traits.Indexer
	traits.Iterable
}

type Map struct {
	mapAccessor
	value any
	size  int
}

func (m *Map) Type() ref.Type {
	return ref.TypeMap
}

func (m *Map) Value() any {
	return m.value
}

func (m *Map) Equal(other ref.Val) bool {
	o, ok := other.(*Map)
	if !ok {
		return false
	}

	if m.size != o.size {
		return false
	}

	mv := reflect.ValueOf(m.value)
	ov := reflect.ValueOf(o.value)
	if mv.Kind() != ov.Kind() {
		return false
	}
	if mv.Kind() == reflect.Map {
		// Objects and arrays are only considered equal when they are the same instance.
		return mv.UnsafePointer() == ov.UnsafePointer()
	}
	return false
}

func (m *Map) Size() int {
	return m.size
}

type mapIterator[K any] struct {
	getter func(K) ref.Val
	keys   []K
	cursor int
}

func (it *mapIterator[K]) HasNext() bool {
	return it.cursor < len(it.keys)
}

func (it *mapIterator[K]) Next() (ref.Val, ref.Val) {
	if it.HasNext() {
		idx := it.keys[it.cursor]
		it.cursor++
		return NativeToVal(idx), it.getter(idx)
	}
	return nil, nil
}
