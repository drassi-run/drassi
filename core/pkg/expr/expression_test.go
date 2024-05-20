package expr_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/dungdm93/drassi/core/pkg/expr"
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
Named values contains value that will be taken out when evaluating expr.
Example of named value: github, job,.... expr will be something like: ${{ github.actor }}

Note that unquoted string literal will be considered named values.
eg: Mona the Octocat is not a string literal.
'Mona the Octocat' is a string literal
*/

func Test_EvaluateLiteral(t *testing.T) {
	type testcase struct {
		expr         string
		expected     any
		expectedKind expr.ResultKind
	}
	tcs := []testcase{
		{"true", true, expr.Boolean},
		{"false", false, expr.Boolean},
		{"711", float64(711), expr.Number},
		{"0", float64(0), expr.Number},
		{"-9.2", -9.2, expr.Number},
		{"0xff", float64(255), expr.Number},
		{"0xaa", float64(170), expr.Number},
		{"-2.99e-2", -0.0299, expr.Number},
		{"1e3", float64(1000), expr.Number},
		{"'It''s open source!'", "It's open source!", expr.String},
		{"'Mona the Octocat'", "Mona the Octocat", expr.String},
	}
	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tc := range tcs {
		t.Run(tc.expr, func(t *testing.T) {
			root := parser.Parse(tc.expr, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
			assert.Equal(t, tc.expectedKind, result.Kind())
		})
	}
}

func Test_EvaluateLogical(t *testing.T) {
	type testcase struct {
		expr         string
		expected     any
		expectedKind expr.ResultKind
	}
	tests := []testcase{
		{"true", true, expr.Boolean},
		{"!true", false, expr.Boolean},
		{"!true && false", false, expr.Boolean},
		{"(1 == 1)", true, expr.Boolean},
		{"(1 == 1) && 2 == 2", true, expr.Boolean},
		{"false || true", true, expr.Boolean},
		{"(1 < 2)", true, expr.Boolean},
		{"(1 != 1)", false, expr.Boolean},
		{"(3 <= 3) || (4 > 5)", true, expr.Boolean},
		{"!((3 > 3) && (4 >= 4))", true, expr.Boolean},
	}
	var namedValues []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	for _, tc := range tests {
		t.Run(tc.expr, func(t *testing.T) {
			root := parser.Parse(tc.expr, namedValues, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
			assert.Equal(t, tc.expectedKind, result.Kind())
		})
	}
}

func Test_EvaluateSimpleFns(t *testing.T) {
	type testcase struct {
		expr         string
		expected     any
		expectedKind expr.ResultKind
	}
	var namedVals []ast.INamedValueInfo[ast.INamedValue]
	var fns []functions.IFnInfo[ast_ifaces.Fn]
	tcs := []testcase{
		{"contains('Hello world', 'llo')", true, expr.Boolean},
		{"startsWith('Hello world', 'He')", true, expr.Boolean},
		{"endsWith('Hello world', 'world')", true, expr.Boolean},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", expr.String},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}", expr.String},
		{"format('Result: {0}', 1 > 2 && 3 > 4)", "Result: false", expr.String},
		{"format('Result: {0}', 1 > 2 || 3 < 4)", "Result: true", expr.String},
	}
	for _, tc := range tcs {
		t.Run(tc.expr, func(t *testing.T) {
			root := parser.Parse(tc.expr, namedVals, fns)
			result, err := evaluator.EvaluateWithDefaults(root, nil, "")
			assert.NilError(t, err)
			assert.Equal(t, tc.expected, result.Value())
			assert.Equal(t, tc.expectedKind, result.Kind())
		})
	}
}

// Test_EvaluateComplexFns are testcase for functions with value depends on from job status, ...
func Test_EvaluateStatusCheckFns(t *testing.T) {
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

// Test_EvaluateNamedValues are testcase where expr are evaluated from input template context, ...
func Test_EvaluateNamedValues(t *testing.T) {
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
