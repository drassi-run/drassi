package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStruct(t *testing.T) {
	t.Run("val", testStructVal)
	t.Run("index", testStructIndex)
	t.Run("iterator", testStructIterator)
}

func testStructVal(t *testing.T) {
	obj := S{}
	s := NewStruct(&obj)

	assert.Equal(t, 6, s.Size())
	assert.Equal(t, &obj, s.Value())
	assert.True(t, s.Equal(NewStruct(obj)))

	for _, v := range valByType {
		assert.False(t, s.Equal(v))
	}
}

func testStructIndex(t *testing.T) {
	//for _, m := range []*Map{
	//	NewMapGeneric(dict),
	//	NewMapDynamic(dict),
	//} {
	//	assert.EqualValues(t, ref.TypeString, m.IndexType())
	//	for k, v := range dict {
	//		e, err := m.Get(k)
	//		assert.NoError(t, err, "error while get key '%s'", k)
	//		assert.EqualValues(t, ref.TypeInteger, e.Type(), "value of key '%s' should be an integer", k)
	//		assert.EqualValues(t, v, e.Value(), "value for key '%s' is not equal", k)
	//	}
	//}
}

func testStructIterator(t *testing.T) {
	//for _, m := range []*Map{
	//	NewMapGeneric(dict),
	//	NewMapDynamic(dict),
	//} {
	//	assert.EqualValues(t, ref.TypeString, m.IndexType())
	//	// track which key is accessed
	//	keys := make(map[string]bool, len(dict))
	//	for k := range dict {
	//		keys[k] = false
	//	}
	//
	//	it := m.Iterator()
	//	for it.HasNext() {
	//		k, v := it.Next()
	//		assert.EqualValues(t, ref.TypeString, k.Type(), "key '%s' should be a string", k)
	//		assert.EqualValues(t, ref.TypeInteger, v.Type(), "value of key '%s' should be an integer", k)
	//		nk := k.Value().(string)
	//		nv := dict[nk]
	//		assert.EqualValues(t, nv, v.Value(), "value for key '%s' is not equal", nk)
	//
	//		keys[nk] = true
	//	}
	//
	//	for k, v := range keys {
	//		assert.True(t, v, "key '%s' is not accessed", k) // All keys are accessed
	//	}
	//}
}
