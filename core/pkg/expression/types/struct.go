package types

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	"golang.org/x/exp/maps"
)

var fieldIndex = make(map[reflect.Type]map[string]int)

type Struct struct {
	object any

	ptr unsafe.Pointer
	val reflect.Value
	typ reflect.Type
}

func NewStruct(object any) *Struct {
	_value := reflect.ValueOf(object)
	var _pointer unsafe.Pointer
	if _value.Kind() == reflect.Pointer {
		_pointer = _value.UnsafePointer()
		_value = _value.Elem()
	} else if _value.Kind() == reflect.Struct {
		r := reflect.ValueOf(&object)
		_pointer = r.UnsafePointer()
	}
	_type := _value.Type()
	buildFieldIndex(_type)

	return &Struct{
		object: object,
		ptr:    _pointer,
		val:    _value,
		typ:    _type,
	}
}

func buildFieldIndex(t reflect.Type) {
	if _, ok := fieldIndex[t]; ok {
		return
	}

	m := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("gha")
		if !ok || len(tag) == 0 {
			continue
		}

		// TODO: not support embedded struct yet
		var key string
		if idx := strings.IndexByte(tag, ','); idx > 0 {
			key = tag[:idx]
		} else {
			key = tag
		}
		m[key] = i
	}

	fieldIndex[t] = m
}

func (s *Struct) Type() ref.Type {
	return ref.TypeStruct
}

func (s *Struct) Value() any {
	return s.object
}

func (s *Struct) Equal(other ref.Val) bool {
	o, ok := other.(*Struct)
	if !ok {
		return false
	}

	if s.val.Kind() != o.val.Kind() {
		return false
	}

	// Objects and arrays are only considered equal when they are the same instance.
	return s.ptr != nil && s.ptr == o.ptr
}

func (s *Struct) Size() int {
	return len(fieldIndex[s.typ])
}

func (s *Struct) IndexType() ref.Type {
	return ref.TypeString
}

func (s *Struct) Get(index any) (ref.Val, error) {
	idx, ok := index.(string)
	if !ok {
		return nil, fmt.Errorf("index must be a string, got %s (%[1]T): %w", index, errInvalidType)
	}

	return s.get(idx), nil
}

func (s *Struct) get(idx string) ref.Val {
	fIdx := fieldIndex[s.typ][idx]
	field := s.val.FieldByIndex([]int{fIdx})
	return NativeToVal(field)
}

func (s *Struct) Iterator() traits.Iterator {
	return &mapIterator[string]{
		getter: s.get,
		keys:   maps.Keys(fieldIndex[s.typ]),
		cursor: 0,
	}
}
