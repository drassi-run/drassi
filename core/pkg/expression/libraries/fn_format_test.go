package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormat(t *testing.T) {
	t.Run("success", testFormatSuccess)
	t.Run("invalid-format", testFormatInvalidFormat)
	t.Run("lazyEval", testFormatLazyEval)
}

func testFormatSuccess(t *testing.T) {
	testcases := []struct {
		inputs   []any
		expected string
	}{
		{[]any{"Hello {0} {1} {2}", "Mona", "the", "Octocat"}, "Hello Mona the Octocat"},
		{[]any{"{0} {1} {2}", "Mona", "the", "Octocat"}, "Mona the Octocat"},
		{[]any{"{{0}} {1} {2}", "Mona", "the", "Octocat"}, "{0} the Octocat"},
		{[]any{"{0} {1} {{2}}", "Mona", "the", "Octocat"}, "Mona the {2}"},
		{[]any{"Hello }}World{{ {0} {1} {2}", "Mona", "the", "Octocat"}, "Hello }World{ Mona the Octocat"},
		{[]any{"你好 from {0} with {1}, 🚀", "δράση", "❤️"}, "你好 from δράση with ❤️, 🚀"},
	}
	for _, tc := range testcases {
		args := toLazies(tc.inputs...)
		res := Format(args[0], args[1:]...)

		verify(t, tc.expected, res, "format(%#q)", tc.inputs)
	}
}

func testFormatInvalidFormat(t *testing.T) {
	testcases := [][]any{
		{"Hello {} world", "Mona", "the", "Octocat"},   // missing arg index
		{"Hello {x} world", "Mona", "the", "Octocat"},  // can't parse int from "x"
		{"Hello { 1} world", "Mona", "the", "Octocat"}, // contains space
		{"Hello {+1} world", "Mona", "the", "Octocat"}, // can't contain sign +/-
		{"Hello {-1} world", "Mona", "the", "Octocat"},
		{"Hello {12{3} world", "Mona", "the", "Octocat"}, // brace inside placeholder
		{"Hello {1}23} world", "Mona", "the", "Octocat"},
		{"Hello {world", "Mona", "the", "Octocat"}, // unclosed brace
		{"Hello }world", "Mona", "the", "Octocat"}, // closing brace without opening
		{"Hello world {0", "Mona", "the", "Octocat"},
		{"Hello {{{ world", "Mona", "the", "Octocat"},
		{"Hello }}} world", "Mona", "the", "Octocat"},
	}
	for _, tc := range testcases {
		args := toLazies(tc...)
		res := Format(args[0], args[1:]...)

		err, _ := res.(error)
		assert.ErrorIs(t, err, errInvalidFormat, tc[0])
	}
}

func testFormatLazyEval(t *testing.T) {
	format := toLazies("Hello {0} {2}")[0]
	args, tracker := toLaziesWithTracker("Mona", "the", "Octocat", "<3")
	_ = Format(format, args...)

	assert.Equal(t, tracker, []bool{true, false, true, false})
}

func toLazy[V any](val V) ref.LazyVal {
	v := types.NativeToVal(val)
	return func() ref.Val {
		return v
	}
}

func toLazies[V any](vals ...V) []ref.LazyVal {
	res := make([]ref.LazyVal, len(vals))
	for i, v := range vals {
		val := types.NativeToVal(v)
		res[i] = func() ref.Val {
			return val
		}
	}

	return res
}

func toLaziesWithTracker[V any](vals ...V) ([]ref.LazyVal, []bool) {
	res := make([]ref.LazyVal, len(vals))
	tracker := make([]bool, len(vals))

	for i, v := range vals {
		idx, val := i, types.NativeToVal(v)
		res[i] = func() ref.Val {
			tracker[idx] = true
			return val
		}
	}

	return res, tracker
}
