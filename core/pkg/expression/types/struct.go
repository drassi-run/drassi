/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"reflect"
	"slices"
	"strings"
	"sync"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
	xtypes "drassi.run/core/util/types"
)

// fieldMap map from field name to index in reflect.Type.Field()
type fieldMap map[string][]int

var (
	fieldIndexMu sync.Mutex
	fieldIndex   = make(map[reflect.Type]fieldMap)
)

type Struct struct {
	object any // MUST be a struct or pointer to struct

	// Kind() MUST be `reflect.Struct`
	val    reflect.Value
	typ    reflect.Type
	fields fieldMap
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

	fields := fieldIndexOf(refTyp)
	return &Struct{
		object: value,
		val:    refVal,
		typ:    refTyp,
		fields: fields,
	}
}

func fieldIndexOf(t reflect.Type) fieldMap {
	fieldIndexMu.Lock()
	defer fieldIndexMu.Unlock()

	if m, ok := fieldIndex[t]; ok {
		return m
	}

	m := make(fieldMap)
	buildFieldIndex(t, nil, m)
	fieldIndex[t] = m
	return m
}

// see json.parseFieldOptions
func buildFieldIndex(t reflect.Type, parentIndex []int, m fieldMap) {
	var inlined []*xtypes.Pair[reflect.Type, []int]

	// Pass 1: Direct fields at current level
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		key, embed := "", false
		if tag, ok := f.Tag.Lookup("json"); ok {
			if idx := strings.IndexByte(tag, ','); idx >= 0 {
				key = tag[:idx]
				opts := strings.Split(tag[idx+1:], ",")
				if slices.Contains(opts, "embed") {
					embed = true
				}
			} else {
				key = tag
				if key == "-" {
					// json:"-" ignores the field
					continue
				}
			}
		}

		curIndex := append(slices.Clone(parentIndex), i)

		// Anonymous struct without explicit tag name, or explicit "embed" tag
		if (f.Anonymous && key == "") || embed {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				inlined = append(inlined, xtypes.NewPair(ft, curIndex))
				continue
			}
		}

		// If no tag or tag has empty name (like json:",omitempty"), default key to field name
		if key == "" {
			key = f.Name
		}

		// Outer fields shadow embedded fields (shallowest depth wins)
		if _, exists := m[key]; !exists {
			m[key] = curIndex
		}
	}

	// Pass 2: Inlined/embedded fields
	for _, ilf := range inlined {
		buildFieldIndex(ilf.Key, ilf.Value, m)
	}
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
	return len(s.fields)
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
	fIdx, ok := s.fields[idx]
	if !ok {
		return NULL
	}

	field, err := s.val.FieldByIndexErr(fIdx)
	if err != nil || !field.IsValid() {
		return NULL
	}
	return NativeToVal(field)
}

func (s *Struct) Items() traits.Iterator {
	return func(yield func(ref.Val, ref.Val) bool) {
		for k, idx := range s.fields {
			field, err := s.val.FieldByIndexErr(idx)
			var val ref.Val
			if err != nil || !field.IsValid() {
				val = NULL
			} else {
				val = NativeToVal(field)
			}
			key := NativeToVal(k)
			if !yield(key, val) {
				return
			}
		}
	}
}

func (s *Struct) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, s.object)
}
