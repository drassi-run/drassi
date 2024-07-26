package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"math"
	"testing"
)

func TestContains(t *testing.T) {
	t.Run("string", testContainsString)
	t.Run("list", testContainsList)
	t.Run("other", testContainsOther)
}

func testContainsString(t *testing.T) {
	// TRUE tests
	tests := [][2]any{
		{"Hello", "ll"},
		{"HELLO", "ll"},
		{"3.141592", 3.14},
		{3.141592, "3.14"},
		{3.141592, 3.14},
		{math.Inf(-1), "inf"},
		{math.Inf(1), "inf"},
		{math.Inf(-1), "infinity"},
		{math.Inf(1), "infinity"},
		{math.NaN(), "nan"},
		{"Avengers: Infinity War", math.Inf(1)},
		{true, "u"},
		{true, "Tr"},
		{nil, ""},
	}

	for _, tt := range tests {
		search := types.NativeToVal(tt[0])
		item := types.NativeToVal(tt[1])
		actual := Contains(search, item)

		verify(t, true, actual, "contains(%v, %v)", tt[0], tt[1])
	}

	// FALSE tests
	tests = [][2]any{
		{"search", "item"},
		{"3.141592", 314},
		{"Avengers: Infinity War", math.Inf(-1)},
		{"Avengers: Infinity War", math.NaN()},
		{3.141592, "314"},
		{3.141592, 314},
		{nil, "null"},
	}

	for _, tt := range tests {
		search := types.NativeToVal(tt[0])
		item := types.NativeToVal(tt[1])
		actual := Contains(search, item)

		verify(t, false, actual, "contains(%v, %v)", tt[0], tt[1])
	}
}

func testContainsList(t *testing.T) {
	type testCase struct {
		list []any
		item any
	}
	// TRUE tests
	tests := []testCase{
		{[]any{"first", "second"}, "first"},
		{[]any{nil, "second"}, ""},
		{[]any{"first", nil, "third"}, ""},
		{[]any{"first", "", "third"}, nil},
		{[]any{3.14, "second"}, "3.14"},
		{[]any{"3.14", "second"}, 3.14},
	}

	for _, tt := range tests {
		search := types.NativeToVal(tt.list)
		item := types.NativeToVal(tt.item)
		actual := Contains(search, item)

		verify(t, true, actual, "contains(%v, %v)", tt.list, tt.item)
	}

	// FALSE tests
	tests = []testCase{
		{[]any{true, "second"}, "true"},
		{[]any{"true", "second"}, true},
		{[]any{"", "second"}, []any{}},
		{[]any{"", "second"}, map[string]any{}},
	}

	for _, tt := range tests {
		search := types.NativeToVal(tt.list)
		item := types.NativeToVal(tt.item)
		actual := Contains(search, item)

		verify(t, false, actual, "contains(%v, %v)", tt.list, tt.item)
	}
}

func testContainsOther(t *testing.T) {
	// Map & Struct search always return false
	for _, s := range []any{mapSS, objectX} {
		search := types.NativeToVal(s)

		for _, i := range logicalRight {
			item := types.NativeToVal(i)
			actual := Contains(search, item)

			verify(t, false, actual, "contains(%v, %v)", s, i)
		}
	}
}
