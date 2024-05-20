package expr_test

import (
	"fmt"
	"math"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/dungdm93/drassi/core/pkg/expr/ast"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/ast_ifaces"
	"github.com/dungdm93/drassi/core/pkg/expr/ast/functions"
	"github.com/dungdm93/drassi/core/pkg/expr/evaluator"
	"github.com/dungdm93/drassi/core/pkg/expr/parser"
	"github.com/dungdm93/drassi/core/pkg/model/contexts"
)

/*
Examples are from https://docs.github.com/en/actions/learn-github-actions/expressions#example-of-literals
Named values are value that was passed to evaluation context state as a meaningful object.
Named values contains value that will be taken out when evaluating input.
Example of named value: github, job,.... input will be something like: ${{ github.actor }}

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
		{"false", false},
		{"711", float64(711)},
		{"0", float64(0)},
		{"-9.2", -9.2},
		{"0xff", float64(255)},
		{"0xaa", float64(170)},
		{"-2.99e-2", -0.0299},
		{"1e3", float64(1000)},
		{"'It''s open source!'", "It's open source!"},
		{"'Mona the Octocat'", "Mona the Octocat"},
		{"true", true},
		{"false", false},
		{"null", nil},
		{"-9.7", -9.7},
		// TODO: int type should be respected
		{"123", 123},
		// TODO: int type should be respected
		{"0xff", 255},
		{"-2.99e-2", -2.99e-2},
		{"'foo'", "foo"},
		{"'it''s foo'", "it's foo"},
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

// See https://github.com/nektos/act/blob/6ce45e3f246f12d6617691de9a2423920d5fdbe6/pkg/exprparser/interpreter_test.go#L114
func TestCompare(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}
	tests := []testcase{
		{"!null", true},
		{"!-10", false},
		{"!0", true},
		{"!3.14", false},
		{"!''", true},
		{"!'abc'", false},
		{"!fromJSON('{}')", false},
		{"!fromJSON('[]')", false},
		{`null == 0`, true},
		{`true == 1`, true},
		{`'' == 0`, true},
		{`'3' == 3`, true},
		{`0 == null`, true},
		{`1 == true`, true},
		{`0 == ''`, true},
		{`3 == '3'`, true},
		{`'TEST' == 'test'`, true},
		{"true > false", true},
		{"true >= false", true},
		{"true >= true", true},
		{"true != false", true},
		{`fromJSON('{}') < 2`, false},
		{`fromJSON('{}') < fromJSON('[]')`, false},
		{`fromJSON('{}') > fromJSON('[]')`, false},
	}

	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			root := parser.Parse(tc.input, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
			fmt.Printf("result.Kind(): %v\n", result.Kind())
		})
	}
}

func TestBooleanEvaluation(t *testing.T) {
	type testcase struct {
		input    string
		expected any
	}
	tests := []testcase{
		{"true", true},
		{"!true", false},
		{"!true && false", false},
		{"(1 == 1)", true},
		{"(1 == 1) && 2 == 2", true},
		{"false || true", true},
		{"(false || (false || true))", true},
		{"(1 < 2)", true},
		{"(1 != 1)", false},
		{"(3 <= 3) || (4 > 5)", true},
		{"!((3 > 3) && (4 >= 4))", true},
		{`'b' >= 'a'`, true},
		{`'a' == 'a'`, true},
		// true &&
		{"true && true", true},
		{"true && false", false},
		{"true && null", nil},
		{"true && -10", -10},
		{"true && 0", 0},
		{"true && 10", 10},
		{"true && 3.14", 3.14},
		{"true && 0.0", 0},
		{"true && Infinity", math.Inf(1)},
		// {"true && -Infinity", math.Inf(-1)},
		{"true && NaN", math.NaN()},
		{"true && ''", ""},
		{"true && 'abc'", "abc"},
		// false &&
		{"false && true", false},
		{"false && false", false},
		{"false && null", false},
		{"false && -10", false},
		{"false && 0", false},
		{"false && 10", false},
		{"false && 3.14", false},
		{"false && 0.0", false},
		{"false && Infinity", false},
		// {"false && -Infinity", false},
		{"false && NaN", false},
		{"false && ''", false},
		{"false && 'abc'", false},
		// true ||
		{"true || true", true},
		{"true || false", true},
		{"true || null", true},
		{"true || -10", true},
		{"true || 0", true},
		{"true || 10", true},
		{"true || 3.14", true},
		{"true || 0.0", true},
		{"true || Infinity", true},
		// {"true || -Infinity", true},
		{"true || NaN", true},
		{"true || ''", true},
		{"true || 'abc'", true},
		// false ||
		{"false || true", true},
		{"false || false", false},
		{"false || null", nil},
		{"false || -10", -10},
		{"false || 0", 0},
		{"false || 10", 10},
		{"false || 3.14", 3.14},
		{"false || 0.0", 0},
		{"false || Infinity", math.Inf(1)},
		// {"false || -Infinity", math.Inf(-1)},
		{"false || NaN", math.NaN()},
		{"false || ''", ""},
		{"false || 'abc'", "abc"},
		// null &&
		{"null && true", nil},
		{"null && false", nil},
		{"null && null", nil},
		{"null && -10", nil},
		{"null && 0", nil},
		{"null && 10", nil},
		{"null && 3.14", nil},
		{"null && 0.0", nil},
		{"null && Infinity", nil},
		// {"null && -Infinity", nil},
		{"null && NaN", nil},
		{"null && ''", nil},
		{"null && 'abc'", nil},
		// null ||
		{"null || true", true},
		{"null || false", false},
		{"null || null", nil},
		{"null || -10", -10},
		{"null || 0", 0},
		{"null || 10", 10},
		{"null || 3.14", 3.14},
		{"null || 0.0", 0},
		{"null || Infinity", math.Inf(1)},
		// {"null || -Infinity", math.Inf(-1)},
		{"null || NaN", math.NaN()},
		{"null || ''", ""},
		{"null || 'abc'", "abc"},
		// -10 &&
		{"-10 && true", true},
		{"-10 && false", false},
		{"-10 && null", nil},
		{"-10 && -10", -10},
		{"-10 && 0", 0},
		{"-10 && 10", 10},
		{"-10 && 3.14", 3.14},
		{"-10 && 0.0", 0},
		{"-10 && Infinity", math.Inf(1)},
		// {"-10 && -Infinity", math.Inf(-1)},
		{"-10 && NaN", math.NaN()},
		{"-10 && ''", ""},
		{"-10 && 'abc'", "abc"},
		// -10 ||
		{"-10 || true", -10},
		{"-10 || false", -10},
		{"-10 || null", -10},
		{"-10 || -10", -10},
		{"-10 || 0", -10},
		{"-10 || 10", -10},
		{"-10 || 3.14", -10},
		{"-10 || 0.0", -10},
		{"-10 || Infinity", -10},
		// {"-10 || -Infinity", -10},
		{"-10 || NaN", -10},
		{"-10 || ''", -10},
		{"-10 || 'abc'", -10},
		// 0 &&
		{"0 && true", 0},
		{"0 && false", 0},
		{"0 && null", 0},
		{"0 && -10", 0},
		{"0 && 0", 0},
		{"0 && 10", 0},
		{"0 && 3.14", 0},
		{"0 && 0.0", 0},
		{"0 && Infinity", 0},
		// {"0 && -Infinity", 0},
		{"0 && NaN", 0},
		{"0 && ''", 0},
		{"0 && 'abc'", 0},
		// 0 ||
		{"0 || true", true},
		{"0 || false", false},
		{"0 || null", nil},
		{"0 || -10", -10},
		{"0 || 0", 0},
		{"0 || 10", 10},
		{"0 || 3.14", 3.14},
		{"0 || 0.0", 0},
		{"0 || Infinity", math.Inf(1)},
		// {"0 || -Infinity", math.Inf(-1)},
		{"0 || NaN", math.NaN()},
		{"0 || ''", ""},
		{"0 || 'abc'", "abc"},
		// 10 &&
		{"10 && true", 10},
		{"10 && false", false},
		{"10 && null", nil},
		{"10 && -10", -10},
		{"10 && 0", 0},
		{"10 && 10", 10},
		{"10 && 3.14", 3.14},
		{"10 && 0.0", 0},
		{"10 && Infinity", math.Inf(1)},
		// {"10 && -Infinity", math.Inf(-1)},
		{"10 && NaN", math.NaN()},
		{"10 && ''", ""},
		{"10 && 'abc'", "abc"},
		// 10 ||
		{"10 || true", 10},
		{"10 || false", 10},
		{"10 || null", 10},
		{"10 || -10", 10},
		{"10 || 0", 10},
		{"10 || 10", 10},
		{"10 || 3.14", 10},
		{"10 || 0.0", 10},
		{"10 || Infinity", 10},
		// {"10 || -Infinity", 10},
		{"10 || NaN", 10},
		{"10 || ''", 10},
		{"10 || 'abc'", 10},
		// 3.14 &&
		{"3.14 && true", 3.14},
		{"3.14 && false", false},
		{"3.14 && null", nil},
		{"3.14 && -10", -10},
		{"3.14 && 0", 0},
		{"3.14 && 10", 10},
		{"3.14 && 3.14", 3.14},
		{"3.14 && 0.0", 0},
		{"3.14 && Infinity", math.Inf(1)},
		// {"3.14 && -Infinity", math.Inf(-1)},
		{"3.14 && NaN", math.NaN()},
		{"3.14 && ''", ""},
		{"3.14 && 'abc'", "abc"},
		// 3.14 ||
		{"3.14 || true", 3.14},
		{"3.14 || false", 3.14},
		{"3.14 || null", 3.14},
		{"3.14 || -10", 3.14},
		{"3.14 || 0", 3.14},
		{"3.14 || 10", 3.14},
		{"3.14 || 3.14", 3.14},
		{"3.14 || 0.0", 3.14},
		{"3.14 || Infinity", 3.14},
		// {"3.14 || -Infinity", 3.14},
		{"3.14 || NaN", 3.14},
		{"3.14 || ''", 3.14},
		{"3.14 || 'abc'", 3.14},
		// Infinity &&
		{"Infinity && true", math.Inf(1)},
		{"Infinity && false", false},
		{"Infinity && null", nil},
		{"Infinity && -10", -10},
		{"Infinity && 0", 0},
		{"Infinity && 10", 10},
		{"Infinity && 3.14", 3.14},
		{"Infinity && 0.0", 0},
		{"Infinity && Infinity", math.Inf(1)},
		// {"Infinity && -Infinity", math.Inf(-1)},
		{"Infinity && NaN", math.NaN()},
		{"Infinity && ''", ""},
		{"Infinity && 'abc'", "abc"},
		// Infinity ||
		{"Infinity || true", math.Inf(1)},
		{"Infinity || false", math.Inf(1)},
		{"Infinity || null", math.Inf(1)},
		{"Infinity || -10", math.Inf(1)},
		{"Infinity || 0", math.Inf(1)},
		{"Infinity || 10", math.Inf(1)},
		{"Infinity || 3.14", math.Inf(1)},
		{"Infinity || 0.0", math.Inf(1)},
		{"Infinity || Infinity", math.Inf(1)},
		// {"Infinity || -Infinity", math.Inf(1)},
		{"Infinity || NaN", math.Inf(1)},
		{"Infinity || ''", math.Inf(1)},
		{"Infinity || 'abc'", math.Inf(1)},
		{"NaN && true", math.NaN()},
		{"NaN && false", math.NaN()},
		{"NaN && null", math.NaN()},
		{"NaN && -10", math.NaN()},
		{"NaN && 0", math.NaN()},
		{"NaN && 10", math.NaN()},
		{"NaN && 3.14", math.NaN()},
		{"NaN && 0.0", math.NaN()},
		{"NaN && Infinity", math.NaN()},
		{"NaN && NaN", math.NaN()},
		{"NaN && ''", math.NaN()},
		{"NaN && 'abc'", math.NaN()},
		{"NaN || true", true},
		{"NaN || false", false},
		{"NaN || null", nil},
		{"NaN || -10", -10},
		{"NaN || 0", 0},
		{"NaN || 10", 10},
		{"NaN || 3.14", 3.14},
		{"NaN || 0.0", 0},
		{"NaN || Infinity", math.Inf(1)},
		{"NaN || NaN", math.NaN()},
		{"NaN || ''", ""},
		{"NaN || 'abc'", "abc"},
		{"'' && true", ""},
		{"'' && false", ""},
		{"'' && null", ""},
		{"'' && -10", ""},
		{"'' && 0", ""},
		{"'' && 10", ""},
		{"'' && 3.14", ""},
		{"'' && 0.0", ""},
		{"'' && Infinity", ""},
		{"'' && NaN", ""},
		{"'' && ''", ""},
		{"'' && 'abc'", ""},
		{"'' || true", true},
		{"'' || false", false},
		{"'' || null", nil},
		{"'' || -10", -10},
		{"'' || 0", 0},
		{"'' || 10", 10},
		{"'' || 3.14", 3.14},
		{"'' || 0.0", 0},
		{"'' || Infinity", math.Inf(1)},
		{"'' || NaN", math.NaN()},
		{"'' || ''", ""},
		{"'' || 'abc'", "abc"},
		{"'abc' && true", true},
		{"'abc' && false", false},
		{"'abc' && null", nil},
		{"'abc' && -10", -10},
		{"'abc' && 0", 0},
		{"'abc' && 10", 10},
		{"'abc' && 3.14", 3.14},
		{"'abc' && 0.0", 0},
		{"'abc' && Infinity", math.Inf(1)},
		{"'abc' && NaN", math.NaN()},
		{"'abc' && ''", ""},
		{"'abc' && 'abc'", "abc"},
		{"'abc' || true", "abc"},
		{"'abc' || false", "abc"},
		{"'abc' || null", "abc"},
		{"'abc' || -10", "abc"},
		{"'abc' || 0", "abc"},
		{"'abc' || 10", "abc"},
		{"'abc' || 3.14", "abc"},
		{"'abc' || 0.0", "abc"},
		{"'abc' || Infinity", "abc"},
		{"'abc' || NaN", "abc"},
		{"'abc' || ''", "abc"},
		{"'abc' || 'abc'", "abc"},
		{"0.0 && true", 0},
		{"-1.5 && true", true},
	}
	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			root := parser.Parse(tc.input, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
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
		{"contains('search', 'item') }}", false, "contains-str-str"},
		{`cOnTaInS('Hello', 'll') }}`, true, "contains-str-casing"},
		{`contains('HELLO', 'll') }}`, true, "contains-str-casing"},
		{`contains('3.141592', 3.14) }}`, true, "contains-str-number"},
		{`contains(3.141592, '3.14') }}`, true, "contains-number-str"},
		{`contains(3.141592, 3.14) }}`, true, "contains-number-number"},
		{`contains(true, 'u') }}`, true, "contains-bool-str"},
		{`contains(null, '') }}`, true, "contains-null-str"},
		{`contains(fromJSON('["first","second"]'), 'first') }}`, true, "contains-item"},
		{`contains(fromJSON('[null,"second"]'), '') }}`, true, "contains-item-null-empty-str"},
		{`contains(fromJSON('["","second"]'), null) }}`, true, "contains-item-empty-str-null"},
		{`contains(fromJSON('[true,"second"]'), 'true') }}`, false, "contains-item-bool-arr"},
		{`contains(fromJSON('["true","second"]'), true) }}`, false, "contains-item-str-bool"},
		{`contains(fromJSON('[3.14,"second"]'), '3.14') }}`, true, "contains-item-number-str"},
		{`contains(fromJSON('[3.14,"second"]'), 3.14) }}`, true, "contains-item-number-number"},
		{`contains(fromJSON('["","second"]'), fromJSON('[]')) }}`, false, "contains-item-str-arr"},
		{`contains(fromJSON('["","second"]'), fromJSON('{}')) }}`, false, "contains-item-str-obj"},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			root := parser.Parse(tt.input, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tc.expected, result.Value())
		})
	}
}


func TestSimpleFns(t *testing.T) {
	type testcase struct {
		expr     string
		expected any
	}
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	tcs := []testcase{
		{"contains('Hello world', 'llo')", true},
		{"startsWith('Hello world', 'He')", true},
		{"endsWith('Hello world', 'world')", true},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat"},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}"},
		{"format('Result: {0}', 1 > 2 && 3 > 4)", "Result: false"},
		{"format('Result: {0}', 1 > 2 || 3 < 4)", "Result: true"},
		// byte array
		{"fromJSON('[0,1]')[1]", 1.0},
	}
	for _, tc := range tcs {
		t.Run(tc.expr, func(t *testing.T) {
			root := parser.Parse(tc.expr, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			fmt.Printf("result.Kind(): %v\n", result.Kind())
			assert.Equal(t, tc.expected, result.Value())
		})
	}
}

// TestComplexFns are testcase for functions with value depends on from job status, ...
func TestStatusCheckFns(t *testing.T) {
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
func TestNamedValues(t *testing.T) {
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
