package types

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestFloat(t *testing.T) {
	t.Run("equals", testFloatEqual)
	t.Run("convert", testFloatConvert)
	t.Run("compare", testFloatCompare)
}

func testFloatEqual(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		for _, f := range []Float{1.23, 0, 1.2e9} {
			assert.True(t, f.Equal(f))
		}
	})

	one := Float(1)
	t.Run("others", func(t *testing.T) {
		for _, v := range valByType {
			assert.False(t, one.Equal(v))
		}
	})

	t.Run("NaN", func(t *testing.T) {
		assert.False(t, NAN.Equal(one))
		assert.False(t, NAN.Equal(NAN))
		assert.False(t, NAN.Equal(POSITIVE_INF))
		assert.False(t, NAN.Equal(NEGATIVE_INF))
	})

	t.Run("Inf", func(t *testing.T) {
		for _, num := range []Float{POSITIVE_INF, NEGATIVE_INF} {
			assert.False(t, num.Equal(one))
			assert.False(t, num.Equal(NAN))
			assert.Equal(t, math.IsInf(float64(num), 1), num.Equal(POSITIVE_INF))
			assert.Equal(t, math.IsInf(float64(num), -1), num.Equal(NEGATIVE_INF))
		}
	})
}

func testFloatConvert(t *testing.T) {
	t.Run("toBoolean", func(t *testing.T) {
		m := map[Float]bool{
			Float(0):     false,
			Float(1):     true,
			Float(0.1):   true,
			Float(1.23):  true,
			NAN:          false,
			POSITIVE_INF: true,
			NEGATIVE_INF: true,
		}

		for v, b := range m {
			assert.Equal(t, b, v.ToBoolean())
		}
	})

	t.Run("toNumber", func(t *testing.T) {
		m := map[Float]float64{
			Float(0):     0,
			Float(1):     1,
			Float(0.1):   0.1,
			Float(1.23):  1.23,
			POSITIVE_INF: math.Inf(1),
			NEGATIVE_INF: math.Inf(-1),
		}

		for v, f := range m {
			assert.Equal(t, f, v.ToNumber())
		}

		assert.True(t, math.IsNaN(NAN.ToNumber()))
	})

	t.Run("toString", func(t *testing.T) {
		m := map[Float]string{
			Float(0):     "0",
			Float(1):     "1",
			Float(2.0):   "2",
			Float(3.):    "3",
			Float(0.1):   "0.1",
			Float(1.23):  "1.23",
			POSITIVE_INF: "Infinity",
			NEGATIVE_INF: "-Infinity",
			NAN:          "NaN",

			Float(0.0001):                "0.0001",
			Float(0.00019):               "0.00019",
			Float(0.000100000000000009):  "0.000100000000000009",
			Float(0.0001000000000000009): "0.000100000000000001",
			Float(0.0001000000000000001): "0.0001",
			Float(0.00001):               "1E-05",
			Float(100000000000000):       "100000000000000",
			Float(1000000000000000):      "1E+15",
			Float(1000000000000001):      "1E+15",

			Float(1.42857e14): "142857000000000",
			Float(1.42857e15): "1.42857E+15",
			Float(1.42857e16): "1.42857E+16",
			Float(1.42857e-1): "0.142857",
			Float(1.42857e-4): "0.000142857",
			Float(1.42857e-5): "1.42857E-05",
		}

		for v, s := range m {
			assert.Equal(t, s, v.ToString())
		}
	})
}

func testFloatCompare(t *testing.T) {
	numbers := []Float{NEGATIVE_INF, Float(-1), Float(0), Float(1), POSITIVE_INF}
	one := Float(1)

	t.Run("vs.Float", func(t *testing.T) {
		for i, num := range numbers {
			for j, num2 := range numbers {
				r, err := num.Compare(num2)
				assert.NoError(t, err)
				if i < j {
					assert.Less(t, r, 0)
				} else if i > j {
					assert.Greater(t, r, 0)
				} else { // i == j
					assert.Equal(t, 0, r)
				}
			}
		}
	})

	t.Run("vs.Val", func(t *testing.T) {
		for name, val := range valByType {
			if _, ok := val.(Float); ok {
				continue
			}

			t.Run(name, func(t *testing.T) {
				_, err := one.Compare(val)
				assert.Error(t, err)
				assert.ErrorIs(t, err, errUncomparable)
			})
		}
	})

	t.Run("vs.NaN", func(t *testing.T) {
		_, err := NAN.Compare(one)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNaNCompare)

		_, err = one.Compare(NAN)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNaNCompare)
	})
}
