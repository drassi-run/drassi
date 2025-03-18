/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestInteger(t *testing.T) {
	t.Run("equals", testIntegerEqual)
	t.Run("convert", testIntegerConvert)
	t.Run("compare", testIntegerCompare)
}

func testIntegerEqual(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		for _, f := range []Integer{7, -7, ONE, ZERO} {
			assert.True(t, f.Equal(f))
		}
	})

	t.Run("others", func(t *testing.T) {
		for _, v := range valByType {
			assert.False(t, ONE.Equal(v))
		}
	})
}

func testIntegerConvert(t *testing.T) {
	negOne := Integer(-1)
	t.Run("toBoolean", func(t *testing.T) {
		m := map[Integer]bool{
			ZERO:   false,
			ONE:    true,
			negOne: true,
		}

		for v, b := range m {
			assert.Equal(t, b, v.ToBoolean())
		}
	})

	t.Run("toNumber", func(t *testing.T) {
		m := map[Integer]float64{
			ZERO:   0,
			ONE:    1,
			negOne: -1,
		}

		for v, f := range m {
			assert.Equal(t, f, v.ToNumber())
		}
	})

	t.Run("toString", func(t *testing.T) {
		m := map[Integer]string{
			ZERO:   "0",
			ONE:    "1",
			negOne: "-1",
		}

		for v, s := range m {
			assert.Equal(t, s, v.ToString())
		}
	})
}

func testIntegerCompare(t *testing.T) {
	// numbers MUST be in order
	numbers := []Integer{math.MinInt32, -1, 0, 1, math.MaxInt32}

	t.Run("vs.Integer", func(t *testing.T) {
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
			if _, ok := val.(Integer); ok {
				continue
			}

			t.Run(name, func(t *testing.T) {
				_, err := ONE.Compare(val)
				assert.Error(t, err)
				assert.ErrorIs(t, err, errUncomparable)
			})
		}
	})
}
