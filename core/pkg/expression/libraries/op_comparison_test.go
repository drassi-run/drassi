package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"math"
	"reflect"
	"testing"
)

func TestCompare(t *testing.T) {
	t.Run("same-type", testCompareSameType)
	t.Run("diff-type", testCompareDiffType)
}

func testCompareSameType(t *testing.T) {
	const (
		equal = 1 << iota
		less
		greater
	)

	type testCase struct {
		left, right any
		com         int
	}

	tests := []testCase{
		{nil, nil, equal},
		{true, true, equal},
		{-1, -1, equal},
		{0, 0, equal},
		{1, 1, equal},
		{math.Inf(-1), math.Inf(-1), equal},
		{-1.0, -1.0, equal},
		{0.0, 0.0, equal},
		{1.0, 1.0, equal},
		{math.Inf(1), math.Inf(1), equal},
		{"foobar", "foobar", equal},
		{"foobar", "FOOBAR", equal}, // ignore case
		{listInt, listInt, equal},
		{mapSS, mapSS, equal},
		{objectX, objectX, equal},

		{false, true, less},
		{-1, 0, less},
		{0, 1, less},
		{math.Inf(-1), -1.0, less},
		{-1.0, 0.0, less},
		{0.0, 1.0, less},
		{0.0, math.Inf(1), less},
		{"foo", "foobar", less},
		{"foo", "FOOBAR", less}, // ignore case
		{"FOO", "foobar", less}, // ignore case
		{"foobar", "fozz", less},
		{"foobar", "FOZZ", less}, // ignore case
		{"FOOBAR", "fozz", less}, // ignore case

		{true, false, greater},
		{0, -1, greater},
		{1, 0, greater},
		{-1.0, math.Inf(-1), greater},
		{0.0, -1.0, greater},
		{1.0, 0.0, greater},
		{math.Inf(1), 0.0, greater},
		{"foobar", "foo", greater},
		{"FOOBAR", "foo", greater}, // ignore case
		{"foobar", "FOO", greater}, // ignore case
		{"fozz", "foobar", greater},
		{"FOZZ", "foobar", greater}, // ignore case
		{"fozz", "FOOBAR", greater}, // ignore case

		// 0 mean can't compare
		{math.NaN(), math.Inf(-1), 0},
		{math.NaN(), -1, 0},
		{math.NaN(), 0, 0},
		{math.NaN(), 1, 0},
		{math.NaN(), math.Inf(1), 0},
		{math.Inf(-1), math.NaN(), 0},
		{-1, math.NaN(), 0},
		{0, math.NaN(), 0},
		{1, math.NaN(), 0},
		{math.Inf(1), math.NaN(), 0},
		{listInt, listFloat, 0},
		{mapSS, mapIS, 0},
		{objectX, objectY, 0},
	}

	for _, tc := range tests {
		lhs, rhs := types.NativeToVal(tc.left), types.NativeToVal(tc.right)

		l := Less(lhs, rhs)
		verify(t, tc.com&less != 0, l, "%v < %v (%[1]T, %[2]T)", tc.left, tc.right)

		le := LessEquals(lhs, rhs)
		verify(t, tc.com&(less|equal) != 0, le, "%v <= %v (%[1]T, %[2]T)", tc.left, tc.right)

		g := Greater(lhs, rhs)
		verify(t, tc.com&greater != 0, g, "%v > %v (%[1]T, %[2]T)", tc.left, tc.right)

		ge := GreaterEquals(lhs, rhs)
		verify(t, tc.com&(greater|equal) != 0, ge, "%v >= %v (%[1]T, %[2]T)", tc.left, tc.right)
	}
}

func testCompareDiffType(t *testing.T) {
	for _, left := range values {
		for _, right := range values {
			if left.value != nil && right.value != nil {
				if reflect.TypeOf(left.value).Kind() == reflect.TypeOf(right.value).Kind() {
					continue
				}
			}

			lhs, rhs := types.NativeToVal(left.value), types.NativeToVal(right.value)

			l := Less(lhs, rhs)
			verify(t, left.numberify < right.numberify, l, "%v < %v (%[1]T, %[2]T)", left.value, right.value)

			le := LessEquals(lhs, rhs)
			verify(t, left.numberify <= right.numberify, le, "%v <= %v (%[1]T, %[2]T)", left.value, right.value)

			g := Greater(lhs, rhs)
			verify(t, left.numberify > right.numberify, g, "%v > %v (%[1]T, %[2]T)", left.value, right.value)

			ge := GreaterEquals(lhs, rhs)
			verify(t, left.numberify >= right.numberify, ge, "%v >= %v (%[1]T, %[2]T)", left.value, right.value)
		}
	}
}
