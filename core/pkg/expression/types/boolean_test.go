package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

var valByType = map[string]ref.Val{
	"boolean": FALSE,
	"integer": Integer(10),
	"float":   Float(1.23),
	"string":  String("hello"),
	"list":    NewListGeneric([]string{"hello", "world"}),
	"map":     NewMapGeneric(map[string]string{"hello": "h", "world": "w"}),
}

func TestBoolean(t *testing.T) {
	t.Run("equals", testBooleanEqual)
	t.Run("convert", testBooleanConvert)
	t.Run("compare", testBooleanCompare)
}

func testBooleanEqual(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		for _, b := range []Boolean{TRUE, FALSE} {
			assert.True(t, b.Equal(b))
		}
	})

	t.Run("others", func(t *testing.T) {
		for _, v := range valByType {
			assert.False(t, TRUE.Equal(v))
		}
	})
}

func testBooleanConvert(t *testing.T) {
	t.Run("toBoolean", func(t *testing.T) {
		m := map[Boolean]bool{
			TRUE:  true,
			FALSE: false,
		}

		for v, b := range m {
			assert.Equal(t, b, v.ToBoolean())
		}
	})

	t.Run("toNumber", func(t *testing.T) {
		m := map[Boolean]float64{
			TRUE:  1,
			FALSE: 0,
		}

		for v, f := range m {
			assert.Equal(t, f, v.ToNumber())
		}
	})

	t.Run("toString", func(t *testing.T) {
		m := map[Boolean]string{
			TRUE:  "true",
			FALSE: "false",
		}

		for v, s := range m {
			assert.Equal(t, s, v.ToString())
		}
	})
}

func testBooleanCompare(t *testing.T) {
	t.Run("vs.Boolean", func(t *testing.T) {
		r, err := TRUE.Compare(TRUE)
		assert.NoError(t, err)
		assert.Equal(t, 0, r)

		r, err = FALSE.Compare(FALSE)
		assert.NoError(t, err)
		assert.Equal(t, 0, r)

		r, err = TRUE.Compare(FALSE)
		assert.NoError(t, err)
		assert.Greater(t, r, 0)

		r, err = FALSE.Compare(TRUE)
		assert.NoError(t, err)
		assert.Less(t, r, 0)
	})

	t.Run("vs.Val", func(t *testing.T) {
		for name, val := range valByType {
			if _, ok := val.(Boolean); ok {
				continue
			}

			t.Run(name, func(t *testing.T) {
				_, err := TRUE.Compare(val)
				assert.Error(t, err)
				assert.ErrorIs(t, err, errUncomparable)
			})
		}
	})
}
