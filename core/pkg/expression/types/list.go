package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

type List struct {
	value  any // MUST be a slice or array
	sizer  func() int
	getter func(int) any
}

func NewListGeneric[L ~[]E, E any](value L) ref.Val {
	return &List{
		value:  value,
		sizer:  func() int { return len(value) },
		getter: func(i int) any { return value[i] },
	}
}

func NewListDynamic(value any) ref.Val {
	refVal := reflect.ValueOf(value)

	if kind := refVal.Kind(); kind != reflect.Slice && kind != reflect.Array {
		return NewError("expect a slice or array, got %T: %w", value, errInvalidType)
	}

	return &List{
		value: value,
		sizer: func() int { return refVal.Len() },
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

	if l.Size() != o.Size() {
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
	return l.sizer()
}

func (l *List) IndexType() ref.Type {
	return ref.TypeInteger
}

func (l *List) Get(index any) ref.Val {
	idx, ok := index.(int)
	if !ok {
		return NewError("index must be an integer, got %s (%[1]T): %w", index, errInvalidType)
	}
	if idx < 0 || idx >= l.Size() {
		return NULL
	}
	return l.get(idx)
}

func (l *List) get(idx int) ref.Val {
	e := l.getter(idx)
	return NativeToVal(e)
}

func (l *List) Iterator() traits.Iterator {
	return func(yield func(ref.Val, ref.Val) bool) {
		for i := 0; i < l.Size(); i++ {
			idx := Integer(i)
			elem := l.get(i)
			if !yield(idx, elem) {
				return
			}
		}
	}
}
