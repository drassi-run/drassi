package types

import (
	"fmt"
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

type List struct {
	value  any
	size   int
	getter func(int) any
}

func NewListGeneric[L ~[]E, E any](elems L) *List {
	return &List{
		value:  elems,
		size:   len(elems),
		getter: func(i int) any { return elems[i] },
	}
}

func NewListDynamic(value any) *List {
	refVal := reflect.ValueOf(value)
	return &List{
		value: value,
		size:  refVal.Len(),
		getter: func(i int) any {
			return refVal.Index(i).Interface()
		},
	}
}

func (l *List) Type() ref.Type {
	return ref.TypeList
}

func (l *List) Value() any {
	return l.value
}

func (l *List) Equal(other ref.Val) bool {
	o, ok := other.(*List)
	if !ok {
		return false
	}

	if l.size != o.size {
		return false
	}

	lv := reflect.ValueOf(l.value)
	ov := reflect.ValueOf(o.value)
	if lv.Kind() != ov.Kind() {
		return false
	}
	switch lv.Kind() {
	case reflect.Slice:
		// Objects and arrays are only considered equal when they are the same instance.
		// NOTE: an exception in Go is empty slice (len = 0 && cap = 0) always point to a same instance.
		return lv.UnsafePointer() == ov.UnsafePointer()
	case reflect.Array:
		// TODO: find a way to compare array pointer in golang
		return false
	default:
		return false
	}
}

func (l *List) Size() int {
	return l.size
}

func (l *List) IndexType() ref.Type {
	return ref.TypeInteger
}

func (l *List) Get(index any) (ref.Val, error) {
	idx, ok := index.(int)
	if !ok {
		return nil, fmt.Errorf("index must be an integer")
	}
	if 0 > idx || idx >= l.size {
		return nil, fmt.Errorf("list index out-of-range, size %d", l.size)
	}
	return l.get(idx), nil
}

func (l *List) get(idx int) ref.Val {
	e := l.getter(idx)
	return NativeToVal(e)
}

func (l *List) Iterator() traits.Iterator {
	return &listIterator{
		list:   l,
		cursor: 0,
		len:    l.size,
	}
}

type listIterator struct {
	list   *List
	cursor int
	len    int
}

func (it *listIterator) HasNext() bool {
	return it.cursor < it.len
}

func (it *listIterator) Next() (ref.Val, ref.Val) {
	if it.HasNext() {
		idx := it.cursor
		it.cursor++
		return Integer(idx), it.list.get(idx)
	}
	return nil, nil
}
