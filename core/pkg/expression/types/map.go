/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"reflect"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

type mapAccessor interface {
	traits.Indexer
	traits.Iterable
	Size() int
}

type Map struct {
	mapAccessor
	value any // MUST be a map
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

	if m.Size() != o.Size() {
		return false
	}

	mv := reflect.ValueOf(m.value)
	ov := reflect.ValueOf(o.value)
	if mv.Kind() == reflect.Map && ov.Kind() == reflect.Map {
		// Objects and arrays are only considered equal when they are the same instance.
		return mv.UnsafePointer() == ov.UnsafePointer()
	}
	return false
}
