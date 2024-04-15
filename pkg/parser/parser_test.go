package parser

import (
	"testing"

	"gotest.tools/v3/assert"

	"expression_parser/runner"
	"expression_parser/runner/mocks"
)

/*
Example are from https://docs.github.com/en/actions/learn-github-actions/expressions#example-of-literals
Note that unquoted string literal will be considered named value
eg: Mona the Octocat is not a string literal.
'Mona the Octocat' is a string literal
*/

func TestEvaluateLiteral(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          LexicalTokenKind
		name          string
	}
	tcs := []testcase{
		{"false", false, LTKBoolean, LTKBoolean.ToString()},
		{"711", float64(711), LTKNumber, LTKNumber.ToString()},
		{"-9.2", -9.2, LTKNumber, LTKNumber.ToString()},
		{"0xff", float64(255), LTKNumber, LTKNumber.ToString()},
		{"-2.99e-2", -0.0299, LTKNumber, LTKNumber.ToString()},
		{"'It''s open source!'", "It's open source!", LTKString, LTKString.ToString()},
		{"'Mona the Octocat'", "Mona the Octocat", LTKString, LTKString.ToString()},
	}
	var namedValues []INamedValueInfo[iNamedValue]
	var fns []IFnInfo[iFn]
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := new(Parser).CreateTree(tc.expression, namedValues, fns)
			result := Evaluate(root, nil, nil, nil, new(EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}

func TestEvaluateLogical(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          LexicalTokenKind
		name          string
	}
	tests := []testcase{
		{"true", true, LTKBoolean, LTKBoolean.ToString()},
		{"!true", false, LTKBoolean, LTKBoolean.ToString()},
		{"!true && false", false, LTKBoolean, LTKBoolean.ToString()},
		{"(1 == 1)", true, LTKBoolean, LTKBoolean.ToString()},
		{"(1 == 1) && 2 == 2", true, LTKBoolean, LTKBoolean.ToString()},
		{"(1 == 1) && 2 == 2", true, LTKBoolean, LTKBoolean.ToString()},
		{"false || true", true, LTKBoolean, LTKBoolean.ToString()},
		{"(1 < 2)", true, LTKBoolean, LTKBoolean.ToString()},
		{"(1 != 1)", false, LTKBoolean, LTKBoolean.ToString()},
		{"(3 <= 3) || (4 > 5)", true, LTKBoolean, LTKBoolean.ToString()},
		{"!((3 > 3) && (4 >= 4))", true, LTKBoolean, LTKBoolean.ToString()},
	}
	var namedValues []INamedValueInfo[iNamedValue]
	var fns []IFnInfo[iFn]
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			root := new(Parser).CreateTree(test.expression, namedValues, fns)
			result := Evaluate(root, nil, nil, nil, new(EvaluationOption))
			assert.Equal(t, test.expectedValue, result.Value())
		})
	}
}

func TestEvaluateSimpleFns(t *testing.T) {
	type testcase struct {
		expression    string
		expectedValue any
		kind          LexicalTokenKind
		name          string
	}
	var namedVals []INamedValueInfo[iNamedValue]
	fns := []IFnInfo[iFn]{
		NewFunctionInfo[AlwaysFn]("always", 0, 2147483647),
	}
	tcs := []testcase{
		{"always()", true, LTKBoolean, LTKBoolean.ToString()},
		{"contains('Hello world', 'llo')", true, LTKBoolean, LTKBoolean.ToString()},
		{"startsWith('Hello world', 'He')", true, LTKBoolean, LTKBoolean.ToString()},
		{"endsWith('Hello world', 'world')", true, LTKBoolean, LTKBoolean.ToString()},
		{"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')", "Hello Mona the Octocat", LTKString, LTKString.ToString()},
		{"format('{{Hello {0} {1} {2}!}}', 'Mona', 'the', 'Octocat')", "{Hello Mona the Octocat!}", LTKString, LTKString.ToString()},
		{"fromJson('{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}')", "{\\\"FAVORITE_FRUIT\\\": \\\"APPLE\\\", \\\"FAVORITE_COLOR\\\": \\\"BLUE\\\"}",
			LTKString, LTKString.ToString()},
	}
	for _, tc := range tcs {
		t.Run(tc.expression, func(t *testing.T) {
			root := new(Parser).CreateTree(tc.expression, namedVals, fns)
			result := Evaluate(root, nil, nil, nil, new(EvaluationOption))
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
	var namedVals []INamedValueInfo[iNamedValue]
	fns := []IFnInfo[iFn]{
		NewFunctionInfo[AlwaysFn]("always", 0, 2147483647),
		NewFunctionInfo[CancelledFn]("cancelled", 0, 0),
		NewFunctionInfo[SuccessFn]("success", 0, 0),
		NewFunctionInfo[FailureFn]("failure", 0, 0),
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
			root := new(Parser).CreateTree(tc.expression, namedVals, fns)
			result := Evaluate(root, nil, nil, mock, new(EvaluationOption))
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
	namedVals := []INamedValueInfo[iNamedValue]{
		NewNamedValueInfo[ContextValueNode]("github"),
	}
	var fns []IFnInfo[iFn]
	// testcases
	tcs := []testCase{
		{
			"evaluate named value", "github.actor", "foo", func() *runner.TemplateContext {
				return &runner.TemplateContext{
					ExpressionValues: map[string]any{
						"github": NewMockGithubContext("foo"),
					},
				}
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			templateCtx := tc.setUpTemplateCtx()
			root := new(Parser).CreateTree(tc.expression, namedVals, fns)
			result := Evaluate(root, nil, nil, templateCtx, new(EvaluationOption))
			assert.Equal(t, tc.expectedValue, result.Value())
		})
	}
}
