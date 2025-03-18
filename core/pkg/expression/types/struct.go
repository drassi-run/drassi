/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"reflect"
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

var fieldIndex = make(map[reflect.Type]map[string]int)

type Struct struct {
	object any // MUST be a struct or pointer to struct

	// Kind() MUST be reflect.Struct
	val reflect.Value
	typ reflect.Type
}

func NewStruct(value any) ref.Val {
	refVal := reflect.ValueOf(value)
	switch refVal.Kind() {
	case reflect.Struct:
		// does nothing
	case reflect.Pointer:
		elem := refVal.Elem()
		if elem.Kind() == reflect.Struct {
			refVal = elem
			break
		}
		fallthrough // error
	default:
		return NewError("%w: expect a struct or pointer to struct, got %T", errInvalidType, value)
	}
	refTyp := refVal.Type()

	buildFieldIndex(refTyp)
	return &Struct{
		object: value,
		val:    refVal,
		typ:    refTyp,
	}
}

func buildFieldIndex(t reflect.Type) {
	if _, ok := fieldIndex[t]; ok {
		return
	}

	m := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("actions")
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

	so := reflect.ValueOf(s.object)
	oo := reflect.ValueOf(o.object)
	if so.Kind() == reflect.Pointer && oo.Kind() == reflect.Pointer {
		// Objects and arrays are only considered equal when they are the same instance.
		return so.UnsafePointer() == oo.UnsafePointer()
	}
	return false
}

func (s *Struct) Size() int {
	return len(fieldIndex[s.typ])
}

func (s *Struct) IndexType() ref.Type {
	return ref.TypeString
}

func (s *Struct) Get(index any) ref.Val {
	idx, ok := index.(string)
	if !ok {
		return NewError("%w: index must be a string, got %T", errInvalidType, index)
	}

	return s.get(idx)
}

func (s *Struct) get(idx string) ref.Val {
	fIdx, ok := fieldIndex[s.typ][idx]
	if !ok {
		return NULL
	}

	field := s.val.Field(fIdx)
	return NativeToVal(field)
}

func (s *Struct) Items() traits.Iterator {
	return func(yield func(ref.Val, ref.Val) bool) {
		for k, idx := range fieldIndex[s.typ] {
			v := s.val.Field(idx)
			key, value := NativeToVal(k), NativeToVal(v)
			if !yield(key, value) {
				return
			}
		}
	}
}
