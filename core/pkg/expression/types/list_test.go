package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestList(t *testing.T) {
	t.Run("val", testListVal)
	t.Run("index", testListIndex)
	t.Run("iterator", testListIterator)
}

var array = []string{"first", "second", "third"}

func testListVal(t *testing.T) {
	for _, l := range []*List{
		NewListGeneric(array),
		NewListDynamic(array),
	} {
		assert.Equal(t, 3, l.Size())
		assert.Equal(t, array, l.Value())
		assert.True(t, l.Equal(NewListGeneric(array)))

		for _, v := range valByType {
			assert.False(t, l.Equal(v))
		}
	}
}

func testListIndex(t *testing.T) {
	for _, l := range []*List{
		NewListGeneric(array),
		NewListDynamic(array),
	} {
		assert.Equal(t, ref.TypeInteger, l.IndexType())
		for i := 0; i < l.Size(); i++ {
			e, err := l.Get(i)
			assert.NoError(t, err, "error while get index '%d'", i)
			assert.Equal(t, ref.TypeString, e.Type(), "value of index '%d' should be a string", i)
			assert.Equal(t, array[i], e.Value(), "value for index '%d' is not equal", i)
		}
	}
}

func testListIterator(t *testing.T) {
	for _, l := range []*List{
		NewListGeneric(array),
		NewListDynamic(array),
	} {
		assert.Equal(t, ref.TypeInteger, l.IndexType())
		i := int64(0)
		it := l.Iterator()
		for it.HasNext() {
			k, v := it.Next()
			assert.Equal(t, ref.TypeInteger, k.Type(), "index '%s' should be an integer", k)
			assert.Equal(t, i, k.Value(), "expect get value in order")
			assert.Equal(t, ref.TypeString, v.Type(), "value of index '%d' should be a string", i)
			assert.Equal(t, array[i], v.Value(), "value for index '%d' is not equal", i)
			i++
		}
		assert.EqualValues(t, len(array), i, "all values should be accessed")
	}
}
