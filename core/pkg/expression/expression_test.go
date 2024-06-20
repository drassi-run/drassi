package expression_test

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/functions"
	"drassi.run/core/pkg/expression/evaluator"
	"drassi.run/core/pkg/expression/parser"
	"drassi.run/core/pkg/model/contexts"
	"gotest.tools/v3/assert"
)

/*
Examples are from:
- https://docs.github.com/en/actions/learn-github-actions/expressions#example-of-literals
- https://github.com/nektos/act/blob/6ce45e3f246f12d6617691de9a2423920d5fdbe6/pkg/exprparser/interpreter_test.go
- https://github.com/nektos/act/blob/ace4cd47c7f099864866b1f60e064fecde7f36ea/pkg/exprparser/functions_test.go
Named values are value that was passed to evaluation context state as a meaningful object.
Named values contains value that will be taken out when evaluating input.
Example of named value: github, job,.... input will be something like: ${{ github.actor

Note that unquoted string literal will be considered named values.
eg: Mona the Octocat is not a string literal.
'Mona the Octocat' is a string literal
*/

func TestLiteral(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}
	tcs := []testcase{
		{"true", true},
		{"null", nil},
		{"false", false},
		{"711", 711},
		// integer larger than int32 will be evaluated to float64. This is consistent with original gha behavior
		// max int64
		{"9223372036854775807", 9.223372036854776e+18},
		{"-9223372036854775807", -9.223372036854776e+18},
		// max int32
		{"2147483647", 2147483647},
		{"2147483647.0", 2147483647},
		// this should fail since original behavior of gha will parse this as 1.23456789012346E+19 (
		// 14 digits after decimal) while go strconv.
		// ParseFloat will parse it as 1.2345678901234567e+1 (
		// 16 digit after decimal) without an option to customize number of digits
		// displayed on exponential
		{"12345678901234567890.0", 1.2345678901234567e+19},
		{"-2147483647", -2147483647},
		{"-10", -10},
		{"Infinity", math.Inf(1)},
		{"0", 0},
		{"0.0", 0},
		{"-9.2", -9.2},
		{"0xff", 255},
		{"0x1f", 0x1f},
		{"-0xaf", -0xaf},
		{"0x0", 0x0},
		{"0xaa", 170},
		{"-2.99e-2", -0.0299},
		{"1e3", 1000},
		{"12e3", 12000},
		{"'It''s open source!'", "It's open source!"},
		{"'Mona the Octocat'", "Mona the Octocat"},
		{"true", true},
		{"false", false},
		{"null", nil},
		{"-9.7", -9.7},
		{"0.567", 0.567},
		{"-1234.567", -1234.567},
		{"123", 123},
		{"0xff", 255},
		{"-2.99e-2", -2.99e-2},
		{"1.2e3", 1200},
		{"-0.123e-12", -0.123e-12},
		{"0e3", 0},
		{"''", ""},
		// {"'こんにちは＼(^o^)／世界😊'", "こんにちは＼(^o^)／世界😊"},
		{"'foo'", "foo"},
		{"'it''s foo'", "it's foo"},
		{"Infinity", math.Inf(1)},
		{"-Infinity", math.Inf(-1)},
	}
	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			root := parser.Parse(tc.input, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
		})
	}
}

func TestCompare(t *testing.T) {
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"!null", true, "not-null"},
		{"!-10", false, "not-neg-num"},
		{"!0", true, "not-zero"},
		{"!3.14", false, "not-pos-float"},
		{"!''", true, "not-empty-str"},
		{"!'abc'", false, "not-str"},
		{"!fromJSON('{}')", false, "not-obj"},
		{"!fromJSON('[]')", false, "not-arr"},
		{`null == 0`, true, "null-coercion"},
		{`true == 1`, true, "boolean-coercion"},
		{`'' == 0`, true, "string-0-coercion"},
		{`'3' == 3`, true, "string-3-coercion"},
		{`10.0 == 10`, true, "float-int-coercion"},
		{`10 == 10.0`, true, "float-int-coercion"},
		{`10 == 10.1`, false, "float-int-coercion"},
		{`10.0 == 10.0`, true, "float-int-coercion"},

		{`0 == null`, true, "null-coercion-alt"},
		{`1 == true`, true, "boolean-coercion-alt"},
		{`0 == ''`, true, "string-0-coercion-alt"},
		{`3 == '3'`, true, "string-3-coercion-alt"},
		{`'TEST' == 'test'`, true, "string-casing"},
		{"true > false", true, "bool-greater-than"},
		{"true >= false", true, "bool-greater-than-eq"},
		{"true >= true", true, "bool-greater-than-1"},
		{"true != false", true, "bool-not-equal"},
		{`fromJSON('{}') < 2`, false, "object-with-less"},
		{`fromJSON('{}') < fromJSON('[]')`, false, "object/arr-with-lt"},
		{`fromJSON('{}') > fromJSON('[]')`, false, "object/arr-with-gt"},
	}

	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tt := range table {
		t.Run(tt.input, func(t *testing.T) {
			root := parser.Parse(tt.input, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tt.expected, result.Value())
			fmt.Printf("result.Kind(): %v\n", result.Kind())
		})
	}
}

func TestBooleanEvaluation(t *testing.T) {
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		// true &&
		{"true && true", true, "true-and"},
		{"true && false", false, "true-and"},
		{"true && null", nil, "true-and"},
		{"true && -10", -10, "true-and"},
		{"true && 0", 0, "true-and"},
		{"true && 10", 10, "true-and"},
		{"true && 3.14", 3.14, "true-and"},
		{"true && 0.0", 0, "true-and"},
		{"true && 10.0", 10, "true-and"},
		{"true && Infinity", math.Inf(1), "true-and"},
		// {"true && -Infinity", math.Inf(-1), "true-and"},
		{"true && NaN", math.NaN(), "true-and"},
		{"true && ''", "", "true-and"},
		{"true && 'abc'", "abc", "true-and"},
		// false &&
		{"false && true", false, "false-and"},
		{"false && false", false, "false-and"},
		{"false && null", false, "false-and"},
		{"false && -10", false, "false-and"},
		{"false && 0", false, "false-and"},
		{"false && 10", false, "false-and"},
		{"false && 3.14", false, "false-and"},
		{"false && 0.0", false, "false-and"},
		{"false && Infinity", false, "false-and"},
		// {"false && -Infinity", false, "false-and"},
		{"false && NaN", false, "false-and"},
		{"false && ''", false, "false-and"},
		{"false && 'abc'", false, "false-and"},
		// true ||
		{"true || true", true, "true-or"},
		{"true || false", true, "true-or"},
		{"true || null", true, "true-or"},
		{"true || -10", true, "true-or"},
		{"true || 0", true, "true-or"},
		{"true || 10", true, "true-or"},
		{"true || 3.14", true, "true-or"},
		{"true || 0.0", true, "true-or"},
		{"true || Infinity", true, "true-or"},
		// {"true || -Infinity", true, "true-or"},
		{"true || NaN", true, "true-or"},
		{"true || ''", true, "true-or"},
		{"true || 'abc'", true, "true-or"},
		// false ||
		{"false || true", true, "false-or"},
		{"false || false", false, "false-or"},
		{"false || null", nil, "false-or"},
		{"false || -10", -10, "false-or"},
		{"false || 0", 0, "false-or"},
		{"false || 10", 10, "false-or"},
		{"false || 3.14", 3.14, "false-or"},
		{"false || 0.0", 0, "false-or"},
		{"false || Infinity", math.Inf(1), "false-or"},
		// {"false || -Infinity", math.Inf(-1), "false-or"},
		{"false || NaN", math.NaN(), "false-or"},
		{"false || ''", "", "false-or"},
		{"false || 'abc'", "abc", "false-or"},
		// null &&
		{"null && true", nil, "null-and"},
		{"null && false", nil, "null-and"},
		{"null && null", nil, "null-and"},
		{"null && -10", nil, "null-and"},
		{"null && 0", nil, "null-and"},
		{"null && 10", nil, "null-and"},
		{"null && 3.14", nil, "null-and"},
		{"null && 0.0", nil, "null-and"},
		{"null && Infinity", nil, "null-and"},
		// {"null && -Infinity", nil, "null-and"},
		{"null && NaN", nil, "null-and"},
		{"null && ''", nil, "null-and"},
		{"null && 'abc'", nil, "null-and"},
		// null ||
		{"null || true", true, "null-or"},
		{"null || false", false, "null-or"},
		{"null || null", nil, "null-or"},
		{"null || -10", -10, "null-or"},
		{"null || 0", 0, "null-or"},
		{"null || 10", 10, "null-or"},
		{"null || 3.14", 3.14, "null-or"},
		{"null || 0.0", 0, "null-or"},
		{"null || Infinity", math.Inf(1), "null-or"},
		// {"null || -Infinity", math.Inf(-1), "null-or"},
		{"null || NaN", math.NaN(), "null-or"},
		{"null || ''", "", "null-or"},
		{"null || 'abc'", "abc", "null-or"},
		// -10 &&
		{"-10 && true", true, "neg-num-and"},
		{"-10 && false", false, "neg-num-and"},
		{"-10 && null", nil, "neg-num-and"},
		{"-10 && -10", -10, "neg-num-and"},
		{"-10 && 0", 0, "neg-num-and"},
		{"-10 && 10", 10, "neg-num-and"},
		{"-10 && 3.14", 3.14, "neg-num-and"},
		{"-10 && 0.0", 0, "neg-num-and"},
		{"-10 && Infinity", math.Inf(1), "neg-num-and"},
		// {"-10 && -Infinity", math.Inf(-1), "neg-num-and"},
		{"-10 && NaN", math.NaN(), "neg-num-and"},
		{"-10 && ''", "", "neg-num-and"},
		{"-10 && 'abc'", "abc", "neg-num-and"},
		// -10 ||
		{"-10 || true", -10, "neg-num-or"},
		{"-10 || false", -10, "neg-num-or"},
		{"-10 || null", -10, "neg-num-or"},
		{"-10 || -10", -10, "neg-num-or"},
		{"-10 || 0", -10, "neg-num-or"},
		{"-10 || 10", -10, "neg-num-or"},
		{"-10 || 3.14", -10, "neg-num-or"},
		{"-10 || 0.0", -10, "neg-num-or"},
		{"-10 || Infinity", -10, "neg-num-or"},
		// {"-10 || -Infinity", -10, "neg-num-or"},
		{"-10 || NaN", -10, "neg-num-or"},
		{"-10 || ''", -10, "neg-num-or"},
		{"-10 || 'abc'", -10, "neg-num-or"},
		// 0 &&
		{"0 && true", 0, "zero-and"},
		{"0 && false", 0, "zero-and"},
		{"0 && null", 0, "zero-and"},
		{"0 && -10", 0, "zero-and"},
		{"0 && 0", 0, "zero-and"},
		{"0 && 10", 0, "zero-and"},
		{"0 && 3.14", 0, "zero-and"},
		{"0 && 0.0", 0, "zero-and"},
		{"0 && Infinity", 0, "zero-and"},
		// {"0 && -Infinity", 0, "zero-and"},
		{"0 && NaN", 0, "zero-and"},
		{"0 && ''", 0, "zero-and"},
		{"0 && 'abc'", 0, "zero-and"},
		// 0 ||
		{"0 || true", true, "zero-or"},
		{"0 || false", false, "zero-or"},
		{"0 || null", nil, "zero-or"},
		{"0 || -10", -10, "zero-or"},
		{"0 || 0", 0, "zero-or"},
		{"0 || 10", 10, "zero-or"},
		{"0 || 3.14", 3.14, "zero-or"},
		{"0 || 0.0", 0, "zero-or"},
		{"0 || Infinity", math.Inf(1), "zero-or"},
		// {"0 || -Infinity", math.Inf(-1), "zero-or"},
		{"0 || NaN", math.NaN(), "zero-or"},
		{"0 || ''", "", "zero-or"},
		{"0 || 'abc'", "abc", "zero-or"},
		// 10 &&
		{"10 && true", true, "pos-num-and"},
		{"10 && false", false, "pos-num-and"},
		{"10 && null", nil, "pos-num-and"},
		{"10 && -10", -10, "pos-num-and"},
		{"10 && 0", 0, "pos-num-and"},
		{"10 && 10", 10, "pos-num-and"},
		{"10 && 3.14", 3.14, "pos-num-and"},
		{"10 && 0.0", 0, "pos-num-and"},
		{"10 && Infinity", math.Inf(1), "pos-num-and"},
		// {"10 && -Infinity", math.Inf(-1), "pos-num-and"},
		{"10 && NaN", math.NaN(), "pos-num-and"},
		{"10 && ''", "", "pos-num-and"},
		{"10 && 'abc'", "abc", "pos-num-and"},
		// 10 ||
		{"10 || true", 10, "pos-num-or"},
		{"10 || false", 10, "pos-num-or"},
		{"10 || null", 10, "pos-num-or"},
		{"10 || -10", 10, "pos-num-or"},
		{"10 || 0", 10, "pos-num-or"},
		{"10 || 10", 10, "pos-num-or"},
		{"10 || 3.14", 10, "pos-num-or"},
		{"10 || 0.0", 10, "pos-num-or"},
		{"10 || Infinity", 10, "pos-num-or"},
		// {"10 || -Infinity", 10, "pos-num-or"},
		{"10 || NaN", 10, "pos-num-or"},
		{"10 || ''", 10, "pos-num-or"},
		{"10 || 'abc'", 10, "pos-num-or"},
		// 3.14 &&
		{"3.14 && true", true, "pos-float-and"},
		{"3.14 && false", false, "pos-float-and"},
		{"3.14 && null", nil, "pos-float-and"},
		{"3.14 && -10", -10, "pos-float-and"},
		{"3.14 && 0", 0, "pos-float-and"},
		{"3.14 && 10", 10, "pos-float-and"},
		{"3.14 && 3.14", 3.14, "pos-float-and"},
		{"3.14 && 0.0", 0, "pos-float-and"},
		{"3.14 && Infinity", math.Inf(1), "pos-float-and"},
		// {"3.14 && -Infinity", math.Inf(-1), "pos-float-and"},
		{"3.14 && NaN", math.NaN(), "pos-float-and"},
		{"3.14 && ''", "", "pos-float-and"},
		{"3.14 && 'abc'", "abc", "pos-float-and"},
		// 3.14 ||
		{"3.14 || true", 3.14, "pos-float-or"},
		{"3.14 || false", 3.14, "pos-float-or"},
		{"3.14 || null", 3.14, "pos-float-or"},
		{"3.14 || -10", 3.14, "pos-float-or"},
		{"3.14 || 0", 3.14, "pos-float-or"},
		{"3.14 || 10", 3.14, "pos-float-or"},
		{"3.14 || 3.14", 3.14, "pos-float-or"},
		{"3.14 || 0.0", 3.14, "pos-float-or"},
		{"3.14 || Infinity", 3.14, "pos-float-or"},
		// {"3.14 || -Infinity", 3.14, "pos-float-or"},
		{"3.14 || NaN", 3.14, "pos-float-or"},
		{"3.14 || ''", 3.14, "pos-float-or"},
		{"3.14 || 'abc'", 3.14, "pos-float-or"},
		// Infinity &&
		{"Infinity && true", true, "pos-inf-and"},
		{"Infinity && false", false, "pos-inf-and"},
		{"Infinity && null", nil, "pos-inf-and"},
		{"Infinity && -10", -10, "pos-inf-and"},
		{"Infinity && 0", 0, "pos-inf-and"},
		{"Infinity && 10", 10, "pos-inf-and"},
		{"Infinity && 3.14", 3.14, "pos-inf-and"},
		{"Infinity && 0.0", 0, "pos-inf-and"},
		{"Infinity && Infinity", math.Inf(1), "pos-inf-and"},
		// {"Infinity && -Infinity", math.Inf(-1), "pos-inf-and"},
		{"Infinity && NaN", math.NaN(), "pos-inf-and"},
		{"Infinity && ''", "", "pos-inf-and"},
		{"Infinity && 'abc'", "abc", "pos-inf-and"},
		// Infinity ||
		{"Infinity || true", math.Inf(1), "pos-inf-or"},
		{"Infinity || false", math.Inf(1), "pos-inf-or"},
		{"Infinity || null", math.Inf(1), "pos-inf-or"},
		{"Infinity || -10", math.Inf(1), "pos-inf-or"},
		{"Infinity || 0", math.Inf(1), "pos-inf-or"},
		{"Infinity || 10", math.Inf(1), "pos-inf-or"},
		{"Infinity || 3.14", math.Inf(1), "pos-inf-or"},
		{"Infinity || 0.0", math.Inf(1), "pos-inf-or"},
		{"Infinity || Infinity", math.Inf(1), "pos-inf-or"},
		// {"Infinity || -Infinity", math.Inf(1), "pos-inf-or"},
		{"Infinity || NaN", math.Inf(1), "pos-inf-or"},
		{"Infinity || ''", math.Inf(1), "pos-inf-or"},
		{"Infinity || 'abc'", math.Inf(1), "pos-inf-or"},
		// -Infinity &&
		{"-Infinity && true", true, "neg-inf-and"},
		{"-Infinity && false", false, "neg-inf-and"},
		{"-Infinity && null", nil, "neg-inf-and"},
		{"-Infinity && -10", -10, "neg-inf-and"},
		{"-Infinity && 0", 0, "neg-inf-and"},
		{"-Infinity && 10", 10, "neg-inf-and"},
		{"-Infinity && 3.14", 3.14, "neg-inf-and"},
		{"-Infinity && 0.0", 0, "neg-inf-and"},
		{"-Infinity && Infinity", math.Inf(1), "neg-inf-and"},
		{"-Infinity && -Infinity", math.Inf(-1), "neg-inf-and"},
		{"-Infinity && NaN", math.NaN(), "neg-inf-and"},
		{"-Infinity && ''", "", "neg-inf-and"},
		{"-Infinity && 'abc'", "abc", "neg-inf-and"},
		// -Infinity ||
		{"-Infinity || true", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || false", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || null", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || -10", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || 0", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || 10", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || 3.14", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || 0.0", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || Infinity", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || -Infinity", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || NaN", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || ''", math.Inf(-1), "neg-inf-or"},
		{"-Infinity || 'abc'", math.Inf(-1), "neg-inf-or"},
		// NaN &&
		{"NaN && true", math.NaN(), "nan-and"},
		{"NaN && false", math.NaN(), "nan-and"},
		{"NaN && null", math.NaN(), "nan-and"},
		{"NaN && -10", math.NaN(), "nan-and"},
		{"NaN && 0", math.NaN(), "nan-and"},
		{"NaN && 10", math.NaN(), "nan-and"},
		{"NaN && 3.14", math.NaN(), "nan-and"},
		{"NaN && 0.0", math.NaN(), "nan-and"},
		{"NaN && Infinity", math.NaN(), "nan-and"},
		// {"NaN && -Infinity", math.NaN(), "nan-and"},
		{"NaN && NaN", math.NaN(), "nan-and"},
		{"NaN && ''", math.NaN(), "nan-and"},
		{"NaN && 'abc'", math.NaN(), "nan-and"},
		// NaN ||
		{"NaN || true", true, "nan-or"},
		{"NaN || false", false, "nan-or"},
		{"NaN || null", nil, "nan-or"},
		{"NaN || -10", -10, "nan-or"},
		{"NaN || 0", 0, "nan-or"},
		{"NaN || 10", 10, "nan-or"},
		{"NaN || 3.14", 3.14, "nan-or"},
		{"NaN || 0.0", 0, "nan-or"},
		{"NaN || Infinity", math.Inf(1), "nan-or"},
		// {"NaN || -Infinity", math.Inf(-1), "nan-or"},
		{"NaN || NaN", math.NaN(), "nan-or"},
		{"NaN || ''", "", "nan-or"},
		{"NaN || 'abc'", "abc", "nan-or"},
		// "" &&
		{"'' && true", "", "empty-str-and"},
		{"'' && false", "", "empty-str-and"},
		{"'' && null", "", "empty-str-and"},
		{"'' && -10", "", "empty-str-and"},
		{"'' && 0", "", "empty-str-and"},
		{"'' && 10", "", "empty-str-and"},
		{"'' && 3.14", "", "empty-str-and"},
		{"'' && 0.0", "", "empty-str-and"},
		{"'' && Infinity", "", "empty-str-and"},
		// {"'' && -Infinity", "", "empty-str-and"},
		{"'' && NaN", "", "empty-str-and"},
		{"'' && ''", "", "empty-str-and"},
		{"'' && 'abc'", "", "empty-str-and"},
		// "" ||
		{"'' || true", true, "empty-str-or"},
		{"'' || false", false, "empty-str-or"},
		{"'' || null", nil, "empty-str-or"},
		{"'' || -10", -10, "empty-str-or"},
		{"'' || 0", 0, "empty-str-or"},
		{"'' || 10", 10, "empty-str-or"},
		{"'' || 3.14", 3.14, "empty-str-or"},
		{"'' || 0.0", 0, "empty-str-or"},
		{"'' || Infinity", math.Inf(1), "empty-str-or"},
		// {"'' || -Infinity", math.Inf(-1), "empty-str-or"},
		{"'' || NaN", math.NaN(), "empty-str-or"},
		{"'' || ''", "", "empty-str-or"},
		{"'' || 'abc'", "abc", "empty-str-or"},
		// "abc" &&
		{"'abc' && true", true, "str-and"},
		{"'abc' && false", false, "str-and"},
		{"'abc' && null", nil, "str-and"},
		{"'abc' && -10", -10, "str-and"},
		{"'abc' && 0", 0, "str-and"},
		{"'abc' && 10", 10, "str-and"},
		{"'abc' && 3.14", 3.14, "str-and"},
		{"'abc' && 0.0", 0, "str-and"},
		{"'abc' && Infinity", math.Inf(1), "str-and"},
		// {"'abc' && -Infinity", math.Inf(-1), "str-and"},
		{"'abc' && NaN", math.NaN(), "str-and"},
		{"'abc' && ''", "", "str-and"},
		{"'abc' && 'abc'", "abc", "str-and"},
		// "abc" ||
		{"'abc' || true", "abc", "str-or"},
		{"'abc' || false", "abc", "str-or"},
		{"'abc' || null", "abc", "str-or"},
		{"'abc' || -10", "abc", "str-or"},
		{"'abc' || 0", "abc", "str-or"},
		{"'abc' || 10", "abc", "str-or"},
		{"'abc' || 3.14", "abc", "str-or"},
		{"'abc' || 0.0", "abc", "str-or"},
		{"'abc' || Infinity", "abc", "str-or"},
		// {"'abc' || -Infinity", "abc", "str-or"},
		{"'abc' || NaN", "abc", "str-or"},
		{"'abc' || ''", "abc", "str-or"},
		{"'abc' || 'abc'", "abc", "str-or"},
		// extra tests
		{"0.0 && true", 0, "float-evaluation-0-alt"},
		{"-1.5 && true", true, "float-evaluation-neg-alt"},
	}

	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tt := range table {
		t.Run(tt.input, func(t *testing.T) {
			root := parser.Parse(tt.input, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			if expected, ok := tt.expected.(float64); ok && math.IsNaN(expected) {
				assert.Equal(t, true, math.IsNaN(result.Value().(float64)))
			} else {
				assert.Equal(t, tt.expected, result.Value())
			}
		})
	}
}

func TestFunctionContains(t *testing.T) {
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"contains('search', 'item')", false, "contains-str-str"},
		{`cOnTaInS('Hello', 'll')`, true, "contains-str-casing"},
		{`contains('HELLO', 'll')`, true, "contains-str-casing"},
		{`contains('3.141592', 3.14)`, true, "contains-str-number"},
		{`contains(3.141592, '3.14')`, true, "contains-number-str"},
		{`contains(3.141592, 3.14)`, true, "contains-number-number"},
		{`contains(true, 'u')`, true, "contains-bool-str"},
		{`contains(null, '')`, true, "contains-null-str"},
		{`contains(fromJSON('["first","second"]'), 'first')`, true, "contains-item"},
		{`contains('123', '')`, true, "contains-item"},
		{`contains(fromJSON('[null,"second"]'), '')`, true, "contains-item-null-empty-str"},
		{`contains(fromJSON('["first","","third"]'), null)`, true, "contains-item-empty-str-null"},
		{`contains(fromJSON('[true,"second"]'), 'true')`, false, "contains-item-bool-arr"},
		{`contains(fromJSON('["true","second"]'), true)`, false, "contains-item-str-bool"},
		{`contains(fromJSON('[3.14,"second"]'), '3.14')`, true, "contains-item-number-str"},
		{`contains(fromJSON('[3.14,"second"]'), 3.14)`, true, "contains-item-number-number"},
		{`contains(fromJSON('["","second"]'), fromJSON('[]'))`, false, "contains-item-str-arr"},
		{`contains(fromJSON('["","second"]'), fromJSON('{}'))`, false, "contains-item-str-obj"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tt.expected, result.Value())
		})
	}
}

func TestFunctionStartsWith(t *testing.T) {
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"startsWith('search', 'se')", true, "startswith-string"},
		{"startsWith('search', 'sa')", false, "startswith-string"},
		{"startsWith('123search', '123s')", true, "startswith-string"},
		{"startsWith(123, 's')", false, "startswith-string"},
		{"startsWith(123, '12')", true, "startswith-string"},
		{"startsWith('123', 12)", true, "startswith-string"},
		{"startsWith(null, '42')", false, "startswith-string"},
		{"startsWith('null', null)", true, "startswith-string"},
		{"startsWith('null', '')", true, "startswith-string"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tt.expected, result.Value())
		})
	}
}

func TestFunctionEndsWith(t *testing.T) {
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"endsWith('search', 'ch')", true, "endsWith-string"},
		{"endsWith('search', 'sa')", false, "endsWith-string"},
		{"endsWith('search123s', '123s')", true, "endsWith-string"},
		{"endsWith(123, 's')", false, "endsWith-string"},
		{"endsWith(123, '23')", true, "endsWith-string"},
		{"endsWith('123', 23)", true, "endsWith-string"},
		{"endsWith(null, '42')", false, "endsWith-string"},
		{"endsWith('null', null)", true, "endsWith-string"},
		{"endsWith('null', '')", true, "endsWith-string"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tt.expected, result.Value())
		})
	}
}

func TestFunctionJoin(t *testing.T) {
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"join(fromJSON('[\"a\", \"b\"]'), ',')", "a,b", "join-arr"},
		{"join('string', ',')", "string", "join-str"},
		{"join(1, ',')", "1", "join-number"},
		{"join(null, ',')", "", "join-number"},
		{"join(fromJSON('[\"a\", \"b\", null]'), null)", "ab", "join-number"},
		{"join(fromJSON('[\"a\", \"b\"]'))", "a,b", "join-number"},
		{"join(fromJSON('[\"a\", \"b\", null]'), 1)", "a1b1", "join-number"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tt.expected, result.Value())
		})
	}
}

func TestFunctionToJson(t *testing.T) {
	namedVals := []ast.INamedValueInfo[ast.INamedValue]{
		ast.NewNamedValueInfo[ast.ContextValueNode]("env"),
	}
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
	}{
		{"toJSON(env)", "{\n  \"key\": \"value\"\n}", "toJSON"},
		{"toJSON(null)", "null", "toJSON-null"},
	}
	ctx := &contexts.Expr{State: &contexts.Context{Env: map[string]string{
		"key": "value",
	}}}
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, ctx, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tt.expected, result.Value())
		})
	}
}

func TestFunctionFormat(t *testing.T) {
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	table := []struct {
		input    string
		expected interface{}
		name     string
		err      error
	}{
		{"format('text')", "text", "format-plain-string", nil},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", "format-plain-string", nil},
		{"format('Hello {0} {1} {2}!', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat!",
			"format-with-placeholders", nil},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}",
			"format-with-escaped-braces", nil},
		{"format('{{0}}', 'test')", "{0}", "format-with-escaped-braces", nil},
		{"format('{{{0}}}', 'test')", "{test}", "format-with-escaped-braces-and-value", nil},
		{"format('}}')", "}", "format-output-closing-brace", nil},
		{`format('Hello "{0}" {1} {2} {3} {4}', null, true, -3.14, NaN, Infinity)`, `Hello "" true -3.14 NaN Infinity`, "format-with-primitives", nil},
		{`format('Hello "{0}" {1} {2}', fromJSON('[0, true, "abc"]'), fromJSON('[{"a":1}]'), fromJSON('{"a":{"b":1}}'))`, `Hello "Array" Array Object`, "format-with-complex-types", nil},
		{"format(true)", "true", "format-with-primitive-args", nil},
		// {"format('echo Hello {0} ${{Test}}', github.undefined_property)", "echo Hello  ${Test}",
		// 	"format-with-undefined-value", parser.ErrorUnrecognizedNamedValue},
		{"format('{0}}', '{1}', 'World')", "Closing bracket without opening one. " +
			"The following format string is invalid: '{0}}'", "format-invalid-format-string", errors.New("invalid format string {0}}")},
		{"format('{0', '{1}', 'World')", "Unclosed brackets. The following format string is invalid: '{0'",
			"format-invalid-format-string", errors.New("invalid format string {0")},
		{"format('{2}', '{1}', 'World')", "", "format-invalid-replacement-reference", evaluator.ErrorInvalidFormatArgIndex},
		{"format('{2147483648}')", "", "format-invalid-replacement-reference", evaluator.ErrorInvalidFormatArgIndex},
		{"format('{0} {1} {2} {3}', 1.0, 1.1, 1234567890.0, 12345678901234567890.0)", "1 1.1 1234567890 1.2345678901234567e+19", "format-floats", nil},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			if tt.err != nil {
				assert.Equal(t, tt.err.Error(), err.Error())
			} else {
				assert.NilError(t, err)
				fmt.Printf("result.Kind(): %v\n", result.Kind())
				assert.Equal(t, tt.expected, result.Value())
			}
		})
	}
}

// TestComplexFns are testcase for functions with value depends on from job status, ...
func TestStatusCheckFunctions(t *testing.T) {
	type testcase struct {
		name        string
		expr        string
		expected    any
		templateCtx func() *contexts.Expr
	}
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]

	// testcase cases
	tcs := []testcase{
		// always()
		{"always()", "always()", true, func() *contexts.Expr {
			return &contexts.Expr{}
		}},
		// cancelled()
		{
			"invoke cancelled() evaluated to true", "cancelled()", true, func() *contexts.Expr {
				return &contexts.Expr{
					State: &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultCancelled,
						},
					},
				}
			},
		},
		{
			"invoke cancelled() evaluated to false", "cancelled()", false, func() *contexts.Expr {
				return &contexts.Expr{
					State: &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultSuccess,
						},
					},
				}
			},
		},
		// success()
		{
			"invoke success() evaluated to true - pre, post and job-level steps", "success()", true,
			func() *contexts.Expr {
				return &contexts.Expr{State: &contexts.Context{
					Job: contexts.Job{
						Status: contexts.ActionResultSuccess,
					},
				},
				}
			},
		},
		{
			"invoke success() evaluated to false - pre, post and job-level steps", "success()", false,
			func() *contexts.Expr {
				return &contexts.Expr{State: &contexts.Context{
					Job: contexts.Job{
						Status: contexts.ActionResultCancelled,
					},
				},
				}
			},
		},
		// failure
		{
			"invoke failure() evaluated to true - pre, post and job-level steps", "failure()", true, func() *contexts.Expr {
				return &contexts.Expr{State: &contexts.Context{
					Job: contexts.Job{
						Status: contexts.ActionResultFailure,
					},
				},
				}
			},
		},
		{
			"invoke failure() evaluated to false - pre, post and job-level steps", "failure()", false,
			func() *contexts.Expr {
				return &contexts.Expr{State: &contexts.Context{
					Job: contexts.Job{
						Status: contexts.ActionResultSuccess,
					},
				},
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tplCtx := tc.templateCtx()
			root := parser.Parse(tc.expr, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, tplCtx, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
		})
	}
}

// TestNamedValues are testcase where input are evaluated from input template context, ...
func TestContexts(t *testing.T) {
	type testCase struct {
		name     string
		expr     string
		expected any
		exprCtx  func() *contexts.Expr
	}
	namedVals := []ast.INamedValueInfo[ast.INamedValue]{
		ast.NewNamedValueInfo[ast.ContextValueNode]("github"),
		ast.NewNamedValueInfo[ast.ContextValueNode]("strategy"),
	}
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	// testcases
	tcs := []testCase{
		{
			"with format", "format('github.actor: {0}', github.actor)", "github.actor: foo", func() *contexts.Expr {
				return &contexts.Expr{
					State: &contexts.Context{
						Github: contexts.Github{
							Actor: "foo",
						},
					},
				}
			},
		},
		{
			"simple",
			"github.actor",
			"foo",
			func() *contexts.Expr {
				return &contexts.Expr{
					State: &contexts.Context{
						Github: contexts.Github{
							Actor: "foo",
						},
					},
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			templateCtx := tc.exprCtx()
			root := parser.Parse(tc.expr, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, templateCtx, "")
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.expected, result.Value())
		})
	}
}
