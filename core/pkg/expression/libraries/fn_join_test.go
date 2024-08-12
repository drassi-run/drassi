package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"math"
	"testing"
)

func TestJoin(t *testing.T) {
	t.Run("string", testJoinString)
	t.Run("list", testJoinList)
	t.Run("other", testJoinOther)
}

var seps = []any{"|", true, false, math.Inf(-1), math.Inf(+1), math.NaN(), -1, 0, 1, 3.14}

func testJoinString(t *testing.T) {
	for _, tc := range values {
		if !tc.primitive {
			continue
		}

		v := types.NativeToVal(tc.value)

		actual := Join(v, nil)
		verify(t, tc.stringify, actual, "join(%v)", tc.value)

		for _, sep := range seps {
			actual = Join(v, toLazy(sep))
			verify(t, tc.stringify, actual, "join(%v, %v)", tc.value, sep)
		}
	}
}

func testJoinList(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		seps := []any{"|", 1}
		type testCase struct {
			list   []any
			nonSep string
			sep    []string
		}
		tests := []testCase{
			{[]any{"a"}, "a", []string{"a", "a"}},
			{[]any{"a", "b"}, "a,b", []string{"a|b", "a1b"}},
			{[]any{"a", "b", "c"}, "a,b,c", []string{"a|b|c", "a1b1c"}},
		}
		for _, tc := range tests {
			v := types.NativeToVal(tc.list)
			actual := Join(v, nil)
			verify(t, tc.nonSep, actual, "join(%v)", tc.list)

			for i, sep := range seps {
				actual = Join(v, toLazy(sep))
				verify(t, tc.sep[i], actual, "join(%v, %v)", tc.list, sep)
			}
		}
	})

	t.Run("all-types", func(t *testing.T) {
		list := make([]any, 0)
		nilSepRes := ""
		sepRes := make([]string, len(seps))
		for i, v := range values {
			list = append(list, v.value)
			if i != 0 {
				nilSepRes += ","
			}
			nilSepRes += v.stringify

			for j, sep := range seps {
				if i != 0 {
					sepRes[j] += stringify(types.NativeToVal(sep))
				}
				sepRes[j] += v.stringify
			}
		}

		v := types.NativeToVal(list)
		actual := Join(v, nil)
		verify(t, nilSepRes, actual, "join(%v)", list)

		for i, sep := range seps {
			actual = Join(v, toLazy(sep))
			verify(t, sepRes[i], actual, "join(%v, %v)", list, sep)
		}
	})
}

func testJoinOther(t *testing.T) {
	for _, o := range []any{mapSS, objectX} {
		v := types.NativeToVal(o)
		actual := Join(v, nil)
		// always return empty string
		verify(t, "", actual, "join(%v)", o)

		for _, sep := range seps {
			actual = Join(v, toLazy(sep))
			verify(t, "", actual, "join(%v, %v)", o, sep)
		}
	}
}
