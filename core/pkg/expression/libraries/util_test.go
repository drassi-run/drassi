/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"math"
	"testing"

	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
)

type weakTypeConversion struct {
	value     any
	truthy    bool
	falsy     bool
	numberify float64
	stringify string
	primitive bool
}

var values = []weakTypeConversion{
	{nil, false, true, 0, "", true},
	{true, true, false, 1, "true", true},
	{false, false, true, 0, "false", true},
	{0, false, true, 0, "0", true},
	{-1, true, false, -1, "-1", true},
	{1, true, false, 1, "1", true},
	{math.Inf(-1), true, false, math.Inf(-1), "-Infinity", true},
	{math.Inf(1), true, false, math.Inf(1), "Infinity", true},
	{math.NaN(), false, true, math.NaN(), "NaN", true},
	{0.0, false, true, 0.0, "0", true},
	{3.14, true, false, 3.14, "3.14", true},
	{-3.14, true, false, -3.14, "-3.14", true},
	{"", false, true, 0, "", true},
	{"0", true, false, 0, "0", true},
	{"-1", true, false, -1, "-1", true},
	{"1", true, false, 1, "1", true},
	{"-Infinity", true, false, math.Inf(-1), "-Infinity", true},
	{"Infinity", true, false, math.Inf(1), "Infinity", true},
	{"NaN", true, false, math.NaN(), "NaN", true},
	{"foobar", true, false, math.NaN(), "foobar", true},
	{listInt, true, false, math.NaN(), "array", false},
	{mapSS, true, false, math.NaN(), "object", false},
	{objectX, true, false, math.NaN(), "object", false},
}

func TestValue(t *testing.T) {
	t.Run("truthy", testValueTruthy)
	t.Run("falsy", testValueFalsy)
	t.Run("numberify", testValueNumberify)
	t.Run("stringify", testValueStringify)
}

func testValueTruthy(t *testing.T) {
	for _, tc := range values {
		v := types.NativeToVal(tc.value)
		actual := isTruthy(v)

		assert.Equal(t, tc.truthy, actual, "%q (%[1]T)", tc.value)
	}
}

func testValueFalsy(t *testing.T) {
	for _, tc := range values {
		v := types.NativeToVal(tc.value)
		actual := isFalsy(v)

		assert.Equal(t, tc.falsy, actual, "%q (%[1]T)", tc.value)
	}
}

func testValueNumberify(t *testing.T) {
	for _, tc := range values {
		v := types.NativeToVal(tc.value)
		actual := coerce(v)

		if math.IsNaN(tc.numberify) {
			assert.True(t, math.IsNaN(actual), "%q (%[1]T)", tc.value)
			continue
		}
		assert.Equal(t, tc.numberify, actual, "%q (%[1]T)", tc.value)
	}
}

func testValueStringify(t *testing.T) {
	for _, tc := range values {
		v := types.NativeToVal(tc.value)
		actual := stringify(v)

		assert.Equal(t, tc.stringify, actual, "%q (%[1]T)", tc.value)
	}
}

func TestFunctionBinding(t *testing.T) {
	t.Run("nullaryFn", func(t *testing.T) {
		fn := nullaryFn(func() ref.Val { return types.String("ok") })
		prog := fn.Bind()
		assert.Equal(t, types.String("ok"), prog())
	})

	t.Run("unaryFn", func(t *testing.T) {
		fn := unaryFn(func(v ref.Val) ref.Val { return types.String("hello " + stringify(v)) })
		prog := fn.Bind(func() ref.Val { return types.String("world") })
		assert.Equal(t, types.String("hello world"), prog())
	})

	t.Run("binaryFn", func(t *testing.T) {
		fn := binaryFn(func(v1, v2 ref.Val) ref.Val { return types.String(stringify(v1) + " " + stringify(v2)) })
		prog := fn.Bind(func() ref.Val { return types.String("foo") }, func() ref.Val { return types.String("bar") })
		assert.Equal(t, types.String("foo bar"), prog())
	})

	t.Run("ternaryFn", func(t *testing.T) {
		fn := ternaryFn(func(v1, v2, v3 ref.Val) ref.Val {
			return types.String(stringify(v1) + "-" + stringify(v2) + "-" + stringify(v3))
		})
		prog := fn.Bind(
			func() ref.Val { return types.String("a") },
			func() ref.Val { return types.String("b") },
			func() ref.Val { return types.String("c") },
		)
		assert.Equal(t, types.String("a-b-c"), prog())
	})
}
