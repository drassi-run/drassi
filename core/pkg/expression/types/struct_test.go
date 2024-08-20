package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStruct(t *testing.T) {
	t.Run("val", testStructVal)
	t.Run("index", testStructIndex)
	t.Run("iterator", testStructIterator)
}

func testStructVal(t *testing.T) {
	t.Run("value", func(t *testing.T) {
		obj := S{}
		s := NewStruct(obj).(*Struct)

		assert.Equal(t, 6, s.Size())
		assert.Equal(t, obj, s.Value()) // compare object value
		assert.False(t, s.Equal(NewStruct(obj)))

		for _, v := range valByType {
			assert.False(t, s.Equal(v))
		}
	})

	t.Run("pointer", func(t *testing.T) {
		obj := &S{}
		s := NewStruct(obj).(*Struct)

		assert.Equal(t, 6, s.Size())
		assert.Equal(t, obj, s.Value()) // compare object value
		assert.True(t, s.Equal(NewStruct(obj)))

		for _, v := range valByType {
			assert.False(t, s.Equal(v))
		}
	})
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
	}
	return obj, m
}

func testStructIndex(t *testing.T) {
	obj, m := fixtureTestStruct()

	for _, s := range []*Struct{
		NewStruct(obj).(*Struct),
		NewStruct(&obj).(*Struct),
	} {
		assert.EqualValues(t, ref.TypeString, s.IndexType())
		for k, v := range m {
			e := s.Get(k)
			err, _ := e.(error)
			assert.NoError(t, err, "error while get key %q", k)
			assert.Equal(t, v, e.Value(), "value for key %q is not equal", k)
		}
	}
}

func testStructIterator(t *testing.T) {
	obj, m := fixtureTestStruct()

	for _, s := range []*Struct{
		NewStruct(obj).(*Struct),
		NewStruct(&obj).(*Struct),
	} {
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
