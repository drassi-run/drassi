/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"testing"

	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	t.Run("val", func(t *testing.T) {
		s := S{}
		t.Run("value", testStructVal(s))
		t.Run("pointer", testStructVal(&s))
	})
	t.Run("index", func(t *testing.T) {
		obj, m := fixtureTestStruct()
		t.Run("value", testStructIndex(obj, m))
		t.Run("pointer", testStructIndex(&obj, m))
	})
	t.Run("iterator", func(t *testing.T) {
		obj, m := fixtureTestStruct()
		t.Run("value", testStructIterator(obj, m))
		t.Run("pointer", testStructIterator(&obj, m))
	})
	t.Run("embedded", testStructEmbedded)
}

func testStructVal[T any](obj T) func(t *testing.T) {
	return func(t *testing.T) {
		s := NewStruct(obj).(*Struct)

		assert.Equal(t, 7, s.Size())
		assert.Equal(t, obj, s.Value()) // compare object value

		for _, v := range valByType {
			assert.False(t, s.Equal(v))
		}
	}
}

func fixtureTestStruct() (S, map[string]any) {
	obj := S{
		Boolean: true,
		Integer: 10,
		Float:   6.93,
		String:  "Drassi",
		List:    []string{"first", "second", "third"},
		Map: map[string]string{
			"first":  "one",
			"second": "two",
			"third":  "three",
		},
		Ignore:   'x',
		EmptyTag: '👍',
	}
	m := map[string]any{
		"boolean": true,
		"integer": int64(10),
		"float":   float64(6.93),
		"string":  "Drassi",
		"list":    []string{"first", "second", "third"},
		"map": map[string]string{
			"first":  "one",
			"second": "two",
			"third":  "three",
		},
		"EmptyTag": int64('👍'),
	}
	return obj, m
}

func testStructIndex[T any](obj T, m map[string]any) func(t *testing.T) {
	return func(t *testing.T) {
		s := NewStruct(obj).(*Struct)

		assert.EqualValues(t, ref.TypeString, s.IndexType())
		for k, v := range m {
			e := s.Get(k)
			err, _ := e.(error)
			assert.NoError(t, err, "error while get key %q", k)
			assert.Equal(t, v, e.Value(), "value for key %q is not equal", k)
		}
	}
}

func testStructIterator[T any](obj T, m map[string]any) func(t *testing.T) {
	return func(t *testing.T) {
		s := NewStruct(obj).(*Struct)
		// track which key is accessed
		keys := make(map[string]bool, len(m))
		for k := range m {
			keys[k] = false
		}

		for k, v := range s.Items() {
			assert.Equal(t, ref.TypeString, k.Type(), "key %q should be a string", k)
			nk := k.Value().(string)
			nv := m[nk]
			assert.EqualValues(t, nv, v.Value(), "value for key %q is not equal", nk)

			keys[nk] = true
		}

		for k, v := range keys {
			assert.True(t, v, "key %q is not accessed", k) // All keys are accessed
		}
	}
}

type EmbeddedBase struct {
	BaseField string `json:"base_field"`
}

type EmbeddedOuter struct {
	EmbeddedBase
	OuterField int `json:"outer_field"`
}

func testStructEmbedded(t *testing.T) {
	obj := &EmbeddedOuter{
		BaseField:  "from_base",
		OuterField: 42,
	}

	s := NewStruct(obj).(*Struct)
	assert.Equal(t, 2, s.Size())
	assert.Equal(t, "from_base", s.Get("base_field").Value())
	assert.Equal(t, int64(42), s.Get("outer_field").Value())
}

func TestStructTagMatrix(t *testing.T) {
	t.Run("no json tag", func(t *testing.T) {
		type NoTagStruct struct {
			FieldOne string
			FieldTwo int
		}
		obj := NoTagStruct{FieldOne: "hello", FieldTwo: 100}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "hello", s.Get("FieldOne").Value())
		assert.Equal(t, int64(100), s.Get("FieldTwo").Value())
	})

	t.Run("default name", func(t *testing.T) {
		type OmitEmptyStruct struct {
			Custom      string `json:"custom,omitempty"`
			DefaultName string `json:",omitempty"`
		}
		obj := OmitEmptyStruct{Custom: "val1", DefaultName: "val2"}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "val1", s.Get("custom").Value())
		assert.Equal(t, "val2", s.Get("DefaultName").Value())
	})

	t.Run("ignored and literal dash", func(t *testing.T) {
		type DashStruct struct {
			Ignored string `json:"-"`
			Dash    string `json:"-,omitempty"`
		}
		obj := DashStruct{Ignored: "skip", Dash: "literal_dash"}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 1, s.Size())
		assert.Equal(t, NULL, s.Get("Ignored"))
		assert.Equal(t, "literal_dash", s.Get("-").Value())
	})

	t.Run("embedded struct with tag name", func(t *testing.T) {
		type SubStruct struct {
			SubField string `json:"sub_field"`
		}
		type ParentStruct struct {
			SubStruct   `json:"named_sub"` // custom name
			ParentField string             `json:"parent_field"`
		}
		obj := ParentStruct{
			SubField:    "inner_val",
			ParentField: "outer_val",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "outer_val", s.Get("parent_field").Value())

		assert.Equal(t, NULL, s.Get("sub_field")) // not inlined!
		subVal := s.Get("named_sub")
		subStruct, ok := subVal.(*Struct)
		assert.True(t, ok)
		assert.Equal(t, "inner_val", subStruct.Get("sub_field").Value())
	})

	t.Run("embedded struct without tag", func(t *testing.T) {
		type SubStruct struct {
			SubField string `json:"sub_field"`
		}
		type ParentStruct struct {
			SubStruct         // w/o json tag
			OuterField string `json:"outer"`
		}
		obj := ParentStruct{
			SubField:   "inner_val",
			OuterField: "outer_val",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "inner_val", s.Get("sub_field").Value())
		assert.Equal(t, "outer_val", s.Get("outer").Value())
	})

	t.Run("embedded struct with empty name", func(t *testing.T) {
		type SubStruct struct {
			SubField string `json:"sub_field"`
		}
		type ParentStruct struct {
			SubStruct  `json:",opt"` // empty name
			OuterField string        `json:"outer"`
		}
		obj := ParentStruct{
			SubField:   "inner_val",
			OuterField: "outer_val",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "inner_val", s.Get("sub_field").Value())
		assert.Equal(t, "outer_val", s.Get("outer").Value())
	})

	t.Run("normal field with embed option", func(t *testing.T) {
		type SubStruct struct {
			SubField string `json:"sub_field"`
		}
		type ParentStruct struct {
			NormalField SubStruct `json:",embed"` // json tag w/ empty name and "embed" option
			OuterField  string    `json:"outer"`
		}
		obj := ParentStruct{
			NormalField: SubStruct{SubField: "inner_val"},
			OuterField:  "outer_val",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "inner_val", s.Get("sub_field").Value())
		assert.Equal(t, "outer_val", s.Get("outer").Value())
	})

	t.Run("shadowed fields", func(t *testing.T) {
		type BaseWithOverlap struct {
			Name  string `json:"name"`
			Extra string `json:"extra"`
		}
		type OverlapParent struct {
			BaseWithOverlap
			Name string `json:"name"`
		}
		obj := OverlapParent{
			BaseWithOverlap: BaseWithOverlap{Name: "inner_name", Extra: "extra_val"},
			Name:            "parent_name",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 2, s.Size())
		assert.Equal(t, "parent_name", s.Get("name").Value()) // shallowest depth wins
		assert.Equal(t, "extra_val", s.Get("extra").Value())
	})

	t.Run("pointer embedded struct", func(t *testing.T) {
		type SubStruct struct {
			SubField string `json:"sub_field"`
		}
		type ParentStruct struct {
			*SubStruct
			Outer string `json:"outer"`
		}

		// Initialized pointer
		objInit := &ParentStruct{
			SubStruct: &SubStruct{SubField: "ptr_val"},
			Outer:     "outer_val",
		}
		sInit := NewStruct(objInit).(*Struct)
		assert.Equal(t, 2, sInit.Size())
		assert.Equal(t, "ptr_val", sInit.Get("sub_field").Value())
		assert.Equal(t, "outer_val", sInit.Get("outer").Value())

		// Nil pointer (should return NULL without panicking)
		objNil := &ParentStruct{
			SubStruct: nil,
			Outer:     "outer_val",
		}
		sNil := NewStruct(objNil).(*Struct)
		assert.Equal(t, 2, sNil.Size())
		assert.Equal(t, NULL, sNil.Get("sub_field"))
		assert.Equal(t, "outer_val", sNil.Get("outer").Value())
	})

	t.Run("deeply nested embed structs", func(t *testing.T) {
		type Level1 struct {
			L1 string `json:"l1"`
		}
		type Level2 struct {
			Level1
			L2 string `json:"l2"`
		}
		type Level3 struct {
			Level2
			L3 string `json:"l3"`
		}
		//goland:noinspection ALL
		obj := &Level3{
			Level2: Level2{
				Level1: Level1{
					L1: "val1",
				},
				L2: "val2",
			},
			L3: "val3",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 3, s.Size())
		assert.Equal(t, "val1", s.Get("l1").Value())
		assert.Equal(t, "val2", s.Get("l2").Value())
		assert.Equal(t, "val3", s.Get("l3").Value())
	})

	t.Run("unexported fields", func(t *testing.T) {
		type UnexportedStruct struct {
			Exported   string `json:"exported"`
			unexported string
		}
		obj := UnexportedStruct{
			Exported:   "visible",
			unexported: "hidden",
		}
		s := NewStruct(obj).(*Struct)
		assert.Equal(t, 1, s.Size())
		assert.Equal(t, "visible", s.Get("exported").Value())
		assert.Equal(t, NULL, s.Get("unexported"))
	})
}
