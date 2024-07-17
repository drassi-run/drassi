package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("val", testMapVal)
	t.Run("index", testMapIndex)
	t.Run("iterator", testMapIterator)
}

var dict = map[string]any{
	"first":  1,
	"second": 2,
	"third":  3,
}

func testMapVal(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict),
		NewMapDynamic(dict),
	} {
		assert.Equal(t, 3, m.Size())
		assert.Equal(t, dict, m.Value())
		assert.True(t, m.Equal(NewMapGeneric(dict)))

		for _, v := range valByType {
			assert.False(t, m.Equal(v))
		}
	}
}

func testMapIndex(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict),
		NewMapDynamic(dict),
	} {
		assert.Equal(t, ref.TypeString, m.IndexType())
		for k, v := range dict {
			e, err := m.Get(k)
			assert.NoError(t, err, "error while get key '%s'", k)
			assert.Equal(t, ref.TypeInteger, e.Type(), "value of key '%s' should be an integer", k)
			assert.EqualValues(t, v, e.Value(), "value for key '%s' is not equal", k)
		}
	}
}

func testMapIterator(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict),
		NewMapDynamic(dict),
	} {
		assert.Equal(t, ref.TypeString, m.IndexType())
		// track which key is accessed
		keys := make(map[string]bool, len(dict))
		for k := range dict {
			keys[k] = false
		}

		it := m.Iterator()
		for it.HasNext() {
			k, v := it.Next()
			assert.Equal(t, ref.TypeString, k.Type(), "key '%s' should be a string", k)
			assert.Equal(t, ref.TypeInteger, v.Type(), "value of key '%s' should be an integer", k)
			nk := k.Value().(string)
			nv := dict[nk]
			assert.EqualValues(t, nv, v.Value(), "value for key '%s' is not equal", nk)

			keys[nk] = true
		}

		for k, v := range keys {
			assert.True(t, v, "key '%s' is not accessed", k) // All keys are accessed
		}
	}
}
