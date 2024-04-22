package expression

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/dungdm93/drasi/pkg/expression/evaluator"
	"github.com/dungdm93/drasi/pkg/expression/parser"
	"github.com/dungdm93/drasi/pkg/expression/parser/functions"
	"github.com/dungdm93/drasi/pkg/runner"
	"github.com/dungdm93/drasi/pkg/runner/mocks"
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

func TestEvaluateLiteral(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          parser.LexicalTokenKind
		name          string
	}
	tcs := []testcase{
		{"false", false, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"711", float64(711), parser.LTKNumber, parser.LTKNumber.ToString()},
		{"-9.2", -9.2, parser.LTKNumber, parser.LTKNumber.ToString()},
		{"0xff", float64(255), parser.LTKNumber, parser.LTKNumber.ToString()},
		{"-2.99e-2", -0.0299, parser.LTKNumber, parser.LTKNumber.ToString()},
		{"'It''s open source!'", "It's open source!", parser.LTKString, parser.LTKString.ToString()},
		{"'Mona the Octocat'", "Mona the Octocat", parser.LTKString, parser.LTKString.ToString()},
	}
	var namedValues []parser.INamedValueInfo[parser.INamedValue]
	var fns []functions.IFnInfo[functions.IFn]
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := new(parser.Parser).CreateTree(tc.expression, namedValues, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}

func TestEvaluateLogical(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          parser.LexicalTokenKind
		name          string
	}
	tests := []testcase{
		{"true", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"!true", false, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"!true && false", false, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(1 == 1)", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(1 == 1) && 2 == 2", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(1 == 1) && 2 == 2", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"false || true", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(1 < 2)", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(1 != 1)", false, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"(3 <= 3) || (4 > 5)", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"!((3 > 3) && (4 >= 4))", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
	}
	var namedValues []parser.INamedValueInfo[parser.INamedValue]
	var fns []functions.IFnInfo[functions.IFn]
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			root := new(parser.Parser).CreateTree(test.expression, namedValues, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, test.expectedValue, result.Value())
		})
	}
}

func TestEvaluateSimpleFns(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          parser.LexicalTokenKind
		name          string
	}
	var namedVals []parser.INamedValueInfo[parser.INamedValue]
	fns := []functions.IFnInfo[functions.IFn]{
		functions.NewFunctionInfo[functions.AlwaysFn]("always", 0, 2147483647),
	}
	tcs := []testcase{
		{"always()", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"contains('Hello world', 'llo')", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"startsWith('Hello world', 'He')", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"endsWith('Hello world', 'world')", true, parser.LTKBoolean, parser.LTKBoolean.ToString()},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", parser.LTKString, parser.LTKString.ToString()},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}", parser.LTKString, parser.LTKString.ToString()},
		{"fromJson('{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}')", "{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}",
			parser.LTKString, parser.LTKString.ToString()},
	}
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := new(parser.Parser).CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, nil, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}

// TestEvaluateComplexFns are testcase for functions with value depend from job status, ...
func TestEvaluateComplexFns(t *testing.T) {
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
		// cancelled()
		{
			"invoke cancelled() evaluated to true", "cancelled()", true, func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("JobContext").Return(&runner.JobContext{Status: runner.ActionResultCancelled}).Once()
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke cancelled() evaluated to false", "cancelled()", false, func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("JobContext").Return(&runner.JobContext{Status: runner.ActionResultSuccess}).Once()
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		// success()
		{
			"invoke success() evaluated to true - composite MAIN step", "success()", true,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultSuccess.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke success() evaluated to false - composite MAIN step", "success()", false,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultCancelled.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke success() evaluated to true - pre, post and job-level steps", "success()", true, func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(false).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("JobContext").Return(&runner.JobContext{Status: runner.ActionResultSuccess}).Once()
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke success() evaluated to false - pre, post and job-level steps", "success()", false,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultFailure.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		// failure()
		{
			"invoke failure() evaluated to true - composite MAIN step", "failure()", true,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultFailure.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke failure() evaluated to false - composite MAIN step", "failure()", false,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultSuccess.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke failure() evaluated to true - pre, post and job-level steps", "failure()", true, func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(false).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("JobContext").Return(&runner.JobContext{Status: runner.ActionResultFailure}).Once()
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
		{
			"invoke failure() evaluated to false - pre, post and job-level steps", "failure()", false,
			func() *runner.TemplateContext {
				exeCtx := new(mocks.IExecutionContext)
				exeCtx.On("IsEmbedded").Return(true).Once()
				exeCtx.On("Stage").Return(runner.ActionRunStageMain).Once()
				exeCtx.On("GetGitHubContext", "action_status").Return(runner.ActionResultSuccess.String())
				return &runner.TemplateContext{State: map[string]any{
					"IExecutionContext": exeCtx,
				}}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			mock := tc.setUpTemplateCtx()
			root := new(parser.Parser).CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, mock, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}

// TestEvaluateNamedValues are testcase where expression are evaluated from input template context, ...
func TestEvaluateNamedValues(t *testing.T) {
	type testCase struct {
		name             string
		expression       string
		expectedValue    any
		setUpTemplateCtx func() *runner.TemplateContext
	}
	namedVals := []parser.INamedValueInfo[parser.INamedValue]{
		parser.NewNamedValueInfo[parser.ContextValueNode]("github"),
	}
	var fns []functions.IFnInfo[functions.IFn]
	// testcases
	tcs := []testCase{
		{
			"EvaluateWithContext named value", "github.actor", "foo", func() *runner.TemplateContext {
				return &runner.TemplateContext{
					ExpressionValues: map[string]any{
						"github": parser.NewMockGithubContext("foo"),
					},
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			templateCtx := tc.setUpTemplateCtx()
			root := new(parser.Parser).CreateTree(tc.expression, namedVals, fns)
			result := evaluator.Evaluate(root, nil, nil, templateCtx, new(evaluator.EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}
