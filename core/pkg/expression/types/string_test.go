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

var unicodeWords = [][2]String{
	{"ΔΡΆΣΗ", "δράση"},           // drassi
	{"ΚΥΒΕΡΝΉΤΗΣ", "κυβερνήτησ"}, // kubernetes
	//{"ΚΥΒΕΡΝΉΤΗΣ", "κυβερνήτης"}, // Uppercase(Σ) = Lowercase(σ,ς)
	{"Đặng Minh Dũng", "đặng minh dũng"},
}

func TestString(t *testing.T) {
	t.Run("equals", testStringEqual)
	t.Run("convert", testStringConvert)
	t.Run("compare", testStringCompare)
}

func testStringEqual(t *testing.T) {
	for _, f := range []String{"", "foobar"} {
		t.Run(f.ToString(), func(t *testing.T) {
			assert.True(t, f.Equal(f))
		})
	}

	t.Run("case-insensitive", func(t *testing.T) {
		strs := []String{"foobar", "FOOBAR", "Foobar", "fooBar", "FooBar"}
		for _, s1 := range strs {
			for _, s2 := range strs {
				assert.True(t, s1.Equal(s2))
			}
		}
	})

	t.Run("unicode", func(t *testing.T) {
		for _, p := range unicodeWords {
			x, y := p[0], p[1]
			assert.True(t, x.Equal(y))
		}
	})

	for k, v := range valByType {
		t.Run(k, func(t *testing.T) {
			assert.False(t, ONE.Equal(v))
		})
	}
}

func testStringConvert(t *testing.T) {
	t.Run("toBoolean", func(t *testing.T) {
		m := map[String]bool{
			"":       false,
			"foobar": true,
		}

		for v, b := range m {
			assert.Equal(t, b, v.ToBoolean())
		}
	})

	t.Run("toNumber", func(t *testing.T) {
		m := map[String]float64{
			"":      0,
			"123":   123,
			"-123":  -123,
			"12.3":  12.3,
			"-12.3": -12.3,
			"3.":    3,
			"-3.":   -3,
			"+3.":   3,
			".14":   0.14,
			"-.14":  -0.14,
			"+.14":  0.14,
			"3e1":   30,
			"3e-1":  0.3,
			"3E1":   30,
			"3E-1":  0.3,
			"3.e1":  30,
			"3.E1":  30,

			// Trim whitespace
			"  123":  123,
			"12.3  ": 12.3,
			" 3e1 ":  30,

			// Octal & hex
			"0x12AF": 0x12AF,
			"0x12af": 0x12AF,
			"0o127":  0o127,

			// large float
			".10000000000000000001": .100000000000000000001,
			".00000000000000000001": 1e-20,
			"100000000000000000000": 1e+20,
			"100000000000000000001": 100000000000000000001,

			// Special cases
			"Infinity":  math.Inf(1),
			"+Infinity": math.Inf(1),
			"-Infinity": math.Inf(-1),
			"NaN":       math.NaN(),
		}

		for v, f := range m {
			n := v.ToNumber()
			if math.IsNaN(f) {
				assert.True(t, math.IsNaN(n), "parse number %v must be NaN", v)
			} else {
				assert.Equal(t, f, v.ToNumber())
			}
		}
	})

	t.Run("toNumber/invalid", func(t *testing.T) {
		s := []String{
			// Octal & hex
			"0X12AF", // X & O MUST be lowercase
			"0O127",
			"-0x12AF", // Not accept sign ±
			"0x12p7",  // go strconv.ParseFloat accept 'p' as exponent char for hex number
			"+0x12af",
			"-0o127",
			"0b01", // Not accept binary

			// Out-of-range int32
			"0x1234567890abcdefABCDEF",
			"0x80000000", // MaxInt32+1
			"0o12345677654321",
			"0o20000000000", // MaxInt32+1

			"Inf",
			"+Inf",
			"-Inf",
			"infinity",
			"+infinity",
			"-infinity",

			// Parse error
			"123_456",  // Not accept underscore "_"
			"12.34.56", // To many .
			"12.3e1e2", // To many exponent parts
			"123E",     // missing exponent after E
			"Hello",
		}
		for _, v := range s {
			n := v.ToNumber()
			assert.True(t, math.IsNaN(n), "parse number %v must be NaN", v)
		}
	})

	t.Run("toString", func(t *testing.T) {
		m := map[String]string{
			"":            "",
			"Hello world": "Hello world",

			// Unicode
			"γεια":     "γεια",
			"привет":   "привет",
			"नमस्ते":   "नमस्ते",
			"Xin chào": "Xin chào",
			"你好":       "你好",
			"안녕":       "안녕",
			"こんにちは":    "こんにちは",

			// emoji
			"👍":  "👍",
			"❤️": "❤️",
		}

		for v, s := range m {
			assert.Equal(t, s, v.ToString())
		}
	})
}

func testStringCompare(t *testing.T) {
	// strings MUST be in order
	strings := []String{"", "AA", "AAA", "aaaa", "bbb", "bbc", "BBd"}
	t.Run("vs.String", func(t *testing.T) {
		for i, num := range strings {
			for j, num2 := range strings {
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

	t.Run("ignore-case", func(t *testing.T) {
		for _, p := range unicodeWords {
			x, y := p[0], p[1]
			r, err := x.Compare(y)
			assert.NoError(t, err)
			assert.Equal(t, 0, r, "compare %q vs %q", x, y)
		}
	})

	s := String("foobar")
	t.Run("vs.Val", func(t *testing.T) {
		for name, val := range valByType {
			if _, ok := val.(String); ok {
				continue
			}

			t.Run(name, func(t *testing.T) {
				_, err := s.Compare(val)
				assert.Error(t, err)
				assert.ErrorIs(t, err, errUncomparable)
			})
		}
	})
}
