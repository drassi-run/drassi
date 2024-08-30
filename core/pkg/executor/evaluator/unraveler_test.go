package evaluator

import (
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	. "drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

var (
	la  = []any{"abc", true, 3.14}
	ls  = []string{"one", "two", "three"}
	ms  = map[string]int{"first": 1, "second": 2, "third": 3}
	mi  = map[int]any{1: "value", 2: 3.14, 3: false}
	env *expression.Env
	ur  *unraveler
)

func init() {
	var err error
	env, err = expression.NewEnv(nil,
		expression.WithLibrary(libraries.StdLib()),
		expression.WithVariable("la", la),
		expression.WithVariable("ls", ls),
		expression.WithVariable("ms", ms),
		expression.WithVariable("mi", mi),
	)
	if err != nil {
		panic(err)
	}

	ur = &unraveler{
		exprCache: make(map[string]*exprCache),
		env:       env,
	}
}

func TestUnravel(t *testing.T) {
	t.Run("literal", testUnravelLiteral)
	t.Run("expression", testUnravelExpression)
	t.Run("sequence", testUnravelSequence)
	t.Run("mapping", testUnravelMapping)
}

func testUnravelLiteral(t *testing.T) {
	var value any = "foobar"
	token := NewLiteralToken(value)
	res, err := token.Unravel(ur)
	assert.Nil(t, err)
	assert.Equal(t, value, res)
}

func testUnravelExpression(t *testing.T) {
	t.Run("success", testUnravelExpressionSuccess)
	t.Run("failed", testUnravelExpressionFailed)
}

func testUnravelExpressionSuccess(t *testing.T) {
	tc := map[string]any{
		"${{ 'foobar' }}":  "foobar",
		"${{ true }}":      true,
		"${{ 123 }}":       123,
		"${{ 3.14 }}":      3.14,
		"${{ -Infinity }}": math.Inf(-1),
		"${{ Infinity }}":  math.Inf(1),
		"${{ NaN }}":       math.NaN(),

		"${{ la }}":                             la,
		"${{ ms }}":                             ms,
		"${{ contains(ls, 'one') }}":            true,
		`${{ fromJson('["first", 3, true]') }}`: []any{"first", float64(3), true},
		`${{ fromJson('{"first":"one","second":2,"third":true}') }}`: map[string]any{
			"first":  "one",
			"second": float64(2),
			"third":  true,
		},

		"${{ 'foobar' }}-abcxyz":  "foobar-abcxyz",
		"${{ true }}-abcxyz":      "true-abcxyz",
		"${{ 123 }}-abcxyz":       "123-abcxyz",
		"${{ 3.14 }}-abcxyz":      "3.14-abcxyz",
		"${{ -Infinity }}-abcxyz": "-Infinity-abcxyz",
		"${{ Infinity }}-abcxyz":  "Infinity-abcxyz",
		"${{ NaN }}-abcxyz":       "NaN-abcxyz",
		"${{ la }}-abcxyz":        "array-abcxyz",
		"${{ ms }}-abcxyz":        "object-abcxyz",
	}

	for expr, expected := range tc {
		token := NewExpressionToken(expr)
		res, err := token.Unravel(ur)
		assert.Nil(t, err, "expression: %s", expr)

		if f, ok := expected.(float64); ok && math.IsNaN(f) {
			a, ok := res.(float64)
			assert.Truef(t, ok && math.IsNaN(a), "expression: %s", expr)
			continue
		}
		assert.EqualValuesf(t, expected, res, "expression: %s", expr)
	}
}

func testUnravelExpressionFailed(t *testing.T) {
	t.Run("parse-error", func(t *testing.T) {
		token := NewExpressionToken("${{ )( }}")
		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "syntax error")
	})

	t.Run("bind-error", func(t *testing.T) {
		token := NewExpressionToken("${{ non_exist_var }}")
		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "undefined variable")
	})

	t.Run("exec-error", func(t *testing.T) {
		token := NewExpressionToken("${{ fromJson('{{{') }}")
		_, err := token.Unravel(ur)
		assert.Error(t, err)
	})
}

func testUnravelSequence(t *testing.T) {
	t.Run("success", testUnravelSequenceSuccess)
	t.Run("failed", testUnravelSequenceFailed)
}

var (
	listToken = NewSequenceToken([]Token{
		NewLiteralToken("string"),
		NewLiteralToken(123),
		NewExpressionToken("${{ la }}"),
		NewLiteralToken(math.Inf(1)),
		NewLiteralToken(true),
		NewExpressionToken("${{ ls }}"),
	})
	listResult = []any{
		"string", 123,
		"abc", true, 3.14, // la
		math.Inf(1), true,
		"one", "two", "three", // ls
	}
)

func testUnravelSequenceSuccess(t *testing.T) {
	res, err := listToken.Unravel(ur)
	assert.Nil(t, err)
	assert.Equal(t, listResult, res)
}

func testUnravelSequenceFailed(t *testing.T) {
	token := NewSequenceToken([]Token{
		NewLiteralToken("string"),
		NewExpressionToken("${{ ) }}"),
		NewExpressionToken("${{ non_exist_var }}"),
	})

	_, err := token.Unravel(ur)
	assert.ErrorContains(t, err, "syntax error")
	assert.ErrorContains(t, err, "undefined variable")
}

func testUnravelMapping(t *testing.T) {
	t.Run("string", testUnravelMappingString)
	t.Run("any", testUnravelMappingAny)
	t.Run("error", testUnravelMappingError)
}

func testUnravelMappingString(t *testing.T) {
	token := NewMappingToken([][2]Token{
		{NewLiteralToken("string"), NewLiteralToken("foobar")},
		{NewLiteralToken("int"), NewLiteralToken(123)},
		{NewLiteralToken("float"), NewLiteralToken(3.14)},
		{NewLiteralToken("bool"), NewLiteralToken(true)},
		{NewExpressionToken("${{ 'expr-key' }}"), NewExpressionToken("${{ 'expr-value' }}")},
		{NewExpressionToken("${{ insert }}"), NewExpressionToken("${{ ms }}")},
	})
	expected := map[string]any{
		"string":   "foobar",
		"int":      123,
		"float":    3.14,
		"bool":     true,
		"expr-key": "expr-value",
		"first":    1, "second": 2, "third": 3, // ms
	}

	res, err := token.Unravel(ur)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

var (
	mapToken = NewMappingToken([][2]Token{
		{NewLiteralToken("string"), NewLiteralToken("foobar")},
		{NewLiteralToken(123), NewLiteralToken(123)},
		{NewLiteralToken(3.14), NewLiteralToken(3.14)},
		{NewLiteralToken(false), NewLiteralToken(true)},
		{NewExpressionToken("${{ 'expr-key' }}"), NewExpressionToken("${{ 'expr-value' }}")},
		{NewExpressionToken("${{ insert }}"), NewExpressionToken("${{ mi }}")},
	})
	mapResult = map[string]any{
		"string":   "foobar",
		"123":      123,
		"3.14":     3.14,
		"false":    true,
		"expr-key": "expr-value",
		"1":        "value", "2": 3.14, "3": false, // mi
	}
)

func testUnravelMappingAny(t *testing.T) {
	res, err := mapToken.Unravel(ur)
	assert.Nil(t, err)
	assert.Equal(t, mapResult, res)
}

func testUnravelMappingError(t *testing.T) {
	t.Run("suberr", func(t *testing.T) {
		token := NewMappingToken([][2]Token{
			{NewLiteralToken("foobar"), NewExpressionToken("${{ ) }}")},
		})

		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "syntax error")
	})

	t.Run("insert-not-map", func(t *testing.T) {
		token := NewMappingToken([][2]Token{
			{NewExpressionToken("${{ insert }}"), NewExpressionToken("${{ 'string' }}")},
		})

		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "can't merge map")
	})

	t.Run("key-not-stringable", func(t *testing.T) {
		token := NewMappingToken([][2]Token{
			{NewExpressionToken("${{ ls }}"), NewExpressionToken("${{ 'string' }}")},
		})

		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "key not a string")
	})

	t.Run("key-duplicate", func(t *testing.T) {
		token := NewMappingToken([][2]Token{
			{NewExpressionToken("${{ 'foobar' }}"), NewLiteralToken("first")},
			{NewExpressionToken("${{ 'foobar' }}"), NewLiteralToken("second")},
		})

		_, err := token.Unravel(ur)
		assert.ErrorContains(t, err, "key duplicate")
	})
}
