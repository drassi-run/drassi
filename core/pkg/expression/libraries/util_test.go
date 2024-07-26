package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

type weakTypeConversion struct {
	value     any
	truthy    bool
	falsy     bool
	numberify float64
	stringify string
}

var values = []weakTypeConversion{
	{nil, false, true, 0, ""},
	{true, true, false, 1, "true"},
	{false, false, true, 0, "false"},
	{0, false, true, 0, "0"},
	{-1, true, false, -1, "-1"},
	{1, true, false, 1, "1"},
	{math.Inf(-1), true, false, math.Inf(-1), "-Infinity"},
	{math.Inf(1), true, false, math.Inf(1), "Infinity"},
	{math.NaN(), false, true, math.NaN(), "NaN"},
	{0.0, false, true, 0.0, "0"},
	{3.14, true, false, 3.14, "3.14"},
	{-3.14, true, false, -3.14, "-3.14"},
	{"", false, true, 0, ""},
	{"0", true, false, 0, "0"},
	{"-1", true, false, -1, "-1"},
	{"1", true, false, 1, "1"},
	{"-Infinity", true, false, math.Inf(-1), "-Infinity"},
	{"Infinity", true, false, math.Inf(1), "Infinity"},
	{"NaN", true, false, math.NaN(), "NaN"},
	{[]any{}, true, false, math.NaN(), "array"},
	{map[string]any{}, true, false, math.NaN(), "object"},
	{&weakTypeConversion{}, true, false, math.NaN(), "object"},
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
