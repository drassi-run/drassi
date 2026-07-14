/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// sample underlying type
type String string
type Boolean bool
type Int int64
type UInt uint64
type Float float64
type List[R any] []R
type Map[K comparable, V any] map[K]V
type Struct struct {
	Name string `actions:"name"`
	Age  int    `actions:"age"`
}

var stringMap = map[any]string{
	nil:           "",
	"string":      "string",
	true:          "true",
	int8(-8):      "-8",
	int16(-16):    "-16",
	int32(-32):    "-32",
	int64(-64):    "-64",
	uint(8):       "8",
	uint16(16):    "16",
	uint32(32):    "32",
	uint64(64):    "64",
	float32(32.5): "32.5",
	float64(64.5): "64.5",
	math.Inf(1):   "Infinity",
	math.Inf(-1):  "-Infinity",
	math.NaN():    "NaN",

	String("string"): "string",
	Boolean(false):   "false",
	Int(123):         "123",
	UInt(123):        "123",
	Float(123.4):     "123.4",
}

func TestStringify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for k, v := range stringMap {
			r, ok := Stringify(k)
			assert.True(t, ok, "Stringify %v (%[1]T)", k)
			assert.Equal(t, v, r, "Stringify %v (%[1]T)", k)
		}
	})

	t.Run("failure", func(t *testing.T) {
		tc := []any{
			[]string{},
			[]int{},
			[]any{},
			map[string]string{},
			map[string]any{},
			map[int]string{},
			map[int]any{},
			map[any]any{},

			List[string]{},
			List[int]{},
			List[any]{},
			Map[string, string]{},
			Map[string, any]{},
			Map[int, string]{},
			Map[int, any]{},
			Map[any, any]{},
			Struct{},
		}
		for _, v := range tc {
			_, ok := Stringify(v)
			assert.False(t, ok, "Stringify %v (%[1]T)", v)
		}
	})
}

func TestListStringify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		l1 := []string{"a", "b", "c"}
		l, ok := ListStringify(l1)
		assert.True(t, ok, "ListStringify %v (%[1]T)", l1)
		assert.Equal(t, l1, l, "ListStringify %v (%[1]T)", l1)

		l2 := List[string](l1)
		l, ok = ListStringify(l2)
		assert.True(t, ok, "ListStringify %v (%[1]T)", l2)
		assert.Equal(t, l1, l, "ListStringify %v (%[1]T)", l2)

		var l3 []any
		var r []string
		for k, v := range stringMap {
			l3 = append(l3, k)
			r = append(r, v)
		}
		l, ok = ListStringify(l3)
		assert.True(t, ok, "ListStringify %v (%[1]T)", l3)
		assert.Equal(t, r, l, "ListStringify %v (%[1]T)", l3)

		l4 := List[any](l3)
		l, ok = ListStringify(l4)
		assert.True(t, ok, "ListStringify %v (%[1]T)", l4)
		assert.Equal(t, r, l, "ListStringify %v (%[1]T)", l4)
	})

	t.Run("failure", func(t *testing.T) {
		for k := range stringMap {
			_, ok := ListStringify(k)
			assert.False(t, ok, "ListStringify %v (%[1]T)", k)
		}

		tc := []any{
			map[string]string{},
			map[string]any{},
			map[int]string{},
			map[int]any{},
			map[any]any{},

			Map[string, string]{},
			Map[string, any]{},
			Map[int, string]{},
			Map[int, any]{},
			Map[any, any]{},
			Struct{},
		}
		for _, v := range tc {
			_, ok := ListStringify(v)
			assert.False(t, ok, "ListStringify %v (%[1]T)", v)
		}
	})
}

func TestMapStringify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m1 := map[string]any{"a": 1, "b": true, "c": "ccc", "d": nil}
		m, ok := MapStringify(m1)
		assert.True(t, ok, "MapStringify %v (%[1]T)", m1)
		assert.Equal(t, m1, m, "MapStringify %v (%[1]T)", m1)

		m2 := Map[string, any](m1)
		m, ok = MapStringify(m2)
		assert.True(t, ok, "MapStringify %v (%[1]T)", m2)
		assert.Equal(t, m1, m, "MapStringify %v (%[1]T)", m2)

		m3 := stringMap
		r := make(map[string]any)
		for k, v := range stringMap {
			s, _ := Stringify(k)
			r[s] = v
		}
		m, ok = MapStringify(m3)
		assert.True(t, ok, "MapStringify %v (%[1]T)", m3)
		assert.Equal(t, r, m, "MapStringify %v (%[1]T)", m3)

		m4 := Map[any, string](m3)
		m, ok = MapStringify(m4)
		assert.True(t, ok, "MapStringify %v (%[1]T)", m4)
		assert.Equal(t, r, m, "MapStringify %v (%[1]T)", m4)
	})

	t.Run("failure", func(t *testing.T) {
		for k := range stringMap {
			_, ok := MapStringify(k)
			assert.False(t, ok, "MapStringify %v (%[1]T)", k)
		}

		tc := []any{
			[]string{},
			[]int{},
			[]any{},

			List[string]{},
			List[int]{},
			List[any]{},
		}
		for _, v := range tc {
			_, ok := MapStringify(v)
			assert.False(t, ok, "MapStringify %v (%[1]T)", v)
		}
	})
}

func TestStructStringify(t *testing.T) {
	s := Struct{Name: "drassi", Age: 1}
	r := map[string]any{"name": "drassi", "age": 1}
	m, ok := StructStringify(s)
	assert.True(t, ok, "StructStringify %v (%[1]T)", s)
	assert.Equal(t, r, m, "StructStringify %v (%[1]T)", s)
}
