package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

type List struct {
	value  any // MUST be a slice or array
	size   int
	getter func(int) any
}

func NewListGeneric[L ~[]E, E any](elems L) ref.Val {
	return &List{
		value:  elems,
		size:   len(elems),
		getter: func(i int) any { return elems[i] },
	}
}

func NewListDynamic(value any) ref.Val {
	refVal := reflect.ValueOf(value)

	if kind := refVal.Kind(); kind != reflect.Slice && kind != reflect.Array {
		return NewError("expect a slice or array, got %T: %w", value, errInvalidType)
	}

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
	if lv.Kind() == reflect.Slice && ov.Kind() == reflect.Slice {
		// Objects and arrays are only considered equal when they are the same instance.
		// NOTE: an exception in Go is empty slice (len = 0 && cap = 0) always point to a same instance.
		return lv.UnsafePointer() == ov.UnsafePointer()
	}
	return false
}

func (l *List) Size() int {
	return l.size
}

func (l *List) IndexType() ref.Type {
	return ref.TypeInteger
}

func (l *List) Get(index any) ref.Val {
	idx, ok := index.(int)
	if !ok {
		return NewError("index must be an integer, got %s (%[1]T): %w", index, errInvalidType)
	}
	if idx < 0 || idx >= l.size {
		return NULL
	}
	return l.get(idx)
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
