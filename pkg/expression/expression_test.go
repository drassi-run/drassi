package expression_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/parser"
	"github.com/dungdm93/drasi/pkg/expression/parser/functions"
	"github.com/dungdm93/drasi/pkg/model/contexts"
	"github.com/dungdm93/drasi/pkg/runner"
)

/*
Example are from https://docs.github.com/en/actions/learn-github-actions/expressions#example-of-literals
Named values are value that was passed to evaluation context state as a meaningful object.
Named values contains value that will be taken out when evaluating expression.
Example of named value: github, job,.... Expression will be something like: ${{ github.actor }}

Note that unquoted string literal will be considered named values.
eg: Mona the Octocat is not a string literal.
'Mona the Octocat' is a string literal
*/

func Test_EvaluateLiteral(t *testing.T) {
	type testcase struct {
		expression        string
		expectedValue     any
		expectedValueKind expression.ValueKind
	}
	tcs := []testcase{
		{"false", false, expression.ValueKindBoolean},
		{"711", float64(711), expression.ValueKindNumber},
		{"-9.2", -9.2, expression.ValueKindNumber},
		{"0xff", float64(255), expression.ValueKindNumber},
		{"-2.99e-2", -0.0299, expression.ValueKindNumber},
		{"'It''s open source!'", "It's open source!", expression.ValueKindString},
		{"'Mona the Octocat'", "Mona the Octocat", expression.ValueKindString},
	}
	var namedValues []parser.INamedValueInfo[parser.INamedValue]
	var fns []functions.IFnInfo[functions.IFn]
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := parser.CreateTree(tc.expression, namedValues, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
			assert.Equal(t, tc.expectedValueKind, result.GetKind())
		})
	}
}

func Test_EvaluateLogical(t *testing.T) {
	type testcase struct {
		expression        string
		expectedValue     any
		expectedValueKind expression.ValueKind
	}
	tests := []testcase{
		{"true", true, expression.ValueKindBoolean},
		{"!true", false, expression.ValueKindBoolean},
		{"!true && false", false, expression.ValueKindBoolean},
		{"(1 == 1)", true, expression.ValueKindBoolean},
		{"(1 == 1) && 2 == 2", true, expression.ValueKindBoolean},
		{"(1 == 1) && 2 == 2", true, expression.ValueKindBoolean},
		{"false || true", true, expression.ValueKindBoolean},
		{"(1 < 2)", true, expression.ValueKindBoolean},
		{"(1 != 1)", false, expression.ValueKindBoolean},
		{"(3 <= 3) || (4 > 5)", true, expression.ValueKindBoolean},
		{"!((3 > 3) && (4 >= 4))", true, expression.ValueKindBoolean},
	}
	var namedValues []parser.INamedValueInfo[parser.INamedValue]
	var fns []functions.IFnInfo[functions.IFn]
	for _, tc := range tests {
		t.Run(tc.expression, func(t *testing.T) {
			root := parser.CreateTree(tc.expression, namedValues, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
			assert.Equal(t, tc.expectedValueKind, result.GetKind())
		})
	}
}

func Test_EvaluateSimpleFns(t *testing.T) {
	type testcase struct {
		expression        string
		expectedValue     any
		expectedValueKind expression.ValueKind
	}
	var namedVals []parser.INamedValueInfo[parser.INamedValue]
	fns := []functions.IFnInfo[functions.IFn]{
		functions.NewFunctionInfo[functions.AlwaysFn]("always", 0, 2147483647),
	}
	tcs := []testcase{
		{"contains('Hello world', 'llo')", true, expression.ValueKindBoolean},
		{"startsWith('Hello world', 'He')", true, expression.ValueKindBoolean},
		{"endsWith('Hello world', 'world')", true, expression.ValueKindBoolean},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", expression.ValueKindString},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}", expression.ValueKindString},
		{"fromJson('{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}')", "{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}",
			expression.ValueKindString},
	}
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := parser.CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
			assert.Equal(t, tc.expectedValueKind, result.GetKind())
		})
	}
}

// Test_EvaluateComplexFns are testcase for functions with value depends on from job status, ...
func Test_EvaluateStatusCheckFns(t *testing.T) {
	type testcase struct {
		name             string
		expression       string
		expectedValue    any
		setUpTemplateCtx func() *runner.TemplateContext
	}
	var namedVals []parser.INamedValueInfo[parser.INamedValue]
	fns := []functions.IFnInfo[functions.IFn]{
		functions.NewFunctionInfo[functions.AlwaysFn]("always", 0, 2147483647),
		functions.NewFunctionInfo[functions.CancelledFn]("cancelled", 0, 0),
		functions.NewFunctionInfo[functions.SuccessFn]("success", 0, 0),
		functions.NewFunctionInfo[functions.FailureFn]("failure", 0, 0),
	}

	// testcase cases
	tcs := []testcase{
		// always()
		{"always()", "always()", true, func() *runner.TemplateContext {
			return &runner.TemplateContext{}
		}},
		// cancelled()
		{
			"invoke cancelled() evaluated to true", "cancelled()", true, func() *runner.TemplateContext {
				return &runner.TemplateContext{
					State: map[string]any{
						"IExecutionContext": &contexts.Context{
							Job: contexts.Job{
								Status: contexts.ActionResultCancelled,
							},
						},
					},
				}
			},
		},
		{
			"invoke cancelled() evaluated to false", "cancelled()", false, func() *runner.TemplateContext {
				return &runner.TemplateContext{
					State: map[string]any{
						"IExecutionContext": &contexts.Context{
							Job: contexts.Job{
								Status: contexts.ActionResultSuccess,
							},
						},
					},
				}
			},
		},
		// success()
		{
			"invoke success() evaluated to true - pre, post and job-level steps", "success()", true,
			func() *runner.TemplateContext {
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultSuccess,
						},
					},
				}}
			},
		},
		{
			"invoke success() evaluated to false - pre, post and job-level steps", "success()", false,
			func() *runner.TemplateContext {
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultCancelled,
						},
					},
				}}
			},
		},
		// failure
		{
			"invoke failure() evaluated to true - pre, post and job-level steps", "failure()", true, func() *runner.TemplateContext {
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultFailure,
						},
					},
				}}
			},
		},
		{
			"invoke failure() evaluated to false - pre, post and job-level steps", "failure()", false,
			func() *runner.TemplateContext {
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": &contexts.Context{
						Job: contexts.Job{
							Status: contexts.ActionResultSuccess,
						},
					},
				}}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tplCtx := tc.setUpTemplateCtx()
			root := parser.CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, tplCtx, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}

// Test_EvaluateNamedValues are testcase where expression are evaluated from input template context, ...
func Test_EvaluateNamedValues(t *testing.T) {
	type testCase struct {
		name             string
		expression       string
		expectedValue    any
		setUpTemplateCtx func() *runner.TemplateContext
	}
	namedVals := []parser.INamedValueInfo[parser.INamedValue]{
		parser.NewNamedValueInfo[parser.ContextValueNode]("github"),
		parser.NewNamedValueInfo[parser.ContextValueNode]("strategy"),
	}
	var fns []functions.IFnInfo[functions.IFn]
	// testcases
	tcs := []testCase{
		{
			"evaluate simple named value", "github.actor", "foo", func() *runner.TemplateContext {
				return &runner.TemplateContext{
					ExpressionValues: map[string]any{
						"github": map[string]any{
							"actor": "foo",
						},
					},
				}
			},
		},
		{
			"evaluate matrix named value", "strategy.matrix", map[string]any{
				"version": []int{10, 12, 14},
				"os":      []string{"ubuntu-latest", "windows-latest"},
			}, func() *runner.TemplateContext {
				return &runner.TemplateContext{
					ExpressionValues: map[string]any{
						"strategy": map[string]any{
							"matrix": map[string]any{
								"version": []int{10, 12, 14},
								"os":      []string{"ubuntu-latest", "windows-latest"},
							},
						},
					},
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			templateCtx := tc.setUpTemplateCtx()
			root := parser.CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, templateCtx, new(evaluator.EvaluationOption))
			assert.DeepEqual(t, tc.expectedValue, result.Value())
		})
	}
}
