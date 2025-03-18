/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ast

import (
	"drassi.run/core/pkg/expression/types"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func defaultOption() Option {
	return Option{
		MaxError:  32,
		MaxDepth:  512,
		MaxLength: 21_000,
	}
}

func TestParseExpression(t *testing.T) {
	t.Run("simple", testParseSimple)
	t.Run("ops-order", testParseOperatorOrder)
	t.Run("error", testParseError)
	t.Run("options", testParseOptions)
}

func testParseSimple(t *testing.T) {
	type testCase struct {
		source string
		node   Node
	}
	tests := []testCase{
		// Literal
		{`'foobar'`, &LiteralNode{Value: types.String(`foobar`)}},
		{`'fo''o"bar'`, &LiteralNode{Value: types.String(`fo'o"bar`)}},
		{`'foo''''bar'`, &LiteralNode{Value: types.String(`foo''bar`)}},
		{`'fo\to\nbar'`, &LiteralNode{Value: types.String(`fo\to\nbar`)}},
		{`0`, &LiteralNode{Value: types.ZERO}},
		{`100`, &LiteralNode{Value: types.Integer(100)}},
		{`3.14e100`, &LiteralNode{Value: types.Float(3.14e100)}},
		{`-2.71828e-5`, &LiteralNode{Value: types.Float(-2.71828e-5)}},
		{`Infinity`, &LiteralNode{Value: types.POSITIVE_INF}},
		{`-Infinity`, &LiteralNode{Value: types.NEGATIVE_INF}},
		{`NaN`, &LiteralNode{Value: types.NAN}},
		{`null`, &LiteralNode{Value: types.NULL}},
		{`true`, &LiteralNode{Value: types.TRUE}},
		{`false`, &LiteralNode{Value: types.FALSE}},

		// Identifier
		{`var`, &VariableNode{"var"}},
		{`foo_bar-bar`, &VariableNode{"foo_bar-bar"}},

		// Operator & Function
		{`!a`, &OperatorNode{"!_", []Node{&VariableNode{"a"}}}},
		{`a > b`, &OperatorNode{"_>_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a >= b`, &OperatorNode{"_>=_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a < b`, &OperatorNode{"_<_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a <= b`, &OperatorNode{"_<=_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a && b`, &OperatorNode{"_&&_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a || b`, &OperatorNode{"_||_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}}},
		{`a.b`, &PropertyAccessNode{&VariableNode{"a"}, []string{"b"}}},
		{`a.b.c`, &PropertyAccessNode{&VariableNode{"a"}, []string{"b", "c"}}},
		{`a[b]`, &IndexAccessNode{&VariableNode{"a"}, []Node{&VariableNode{"b"}}}},
		{`a[b][c]`, &IndexAccessNode{&VariableNode{"a"}, []Node{&VariableNode{"b"}, &VariableNode{"c"}}}},
		{`foo()`, &FunctionNode{"foo", []Node{}}},
		{`foo(bar, 'baz')`, &FunctionNode{"foo", []Node{&VariableNode{"bar"}, &LiteralNode{Value: types.String(`baz`)}}}},
	}
	for _, tc := range tests {
		node, err := Parse(tc.source, true, defaultOption())

		assert.NoError(t, err)
		if expected, ok := tc.node.(*LiteralNode); ok && types.IsNaN(expected.Value) {
			actual, ok := node.(*LiteralNode)
			assert.True(t, ok && types.IsNaN(actual.Value))
			continue
		}
		assert.EqualValues(t, tc.node, node, tc.source)
	}
}

func testParseOperatorOrder(t *testing.T) {
	type testCase struct {
		source string
		node   Node
	}
	tests := []testCase{
		{`!a || b && c`, &OperatorNode{
			"_||_", []Node{
				&OperatorNode{"!_", []Node{&VariableNode{"a"}}},
				&OperatorNode{"_&&_", []Node{&VariableNode{"b"}, &VariableNode{"c"}}},
			},
		}},
		{`(!a || b) && c`, &OperatorNode{
			"_&&_", []Node{
				&OperatorNode{"_||_", []Node{
					&OperatorNode{"!_", []Node{&VariableNode{"a"}}},
					&VariableNode{"b"},
				}},
				&VariableNode{"c"},
			},
		}},
		{`a > b || c < d`, &OperatorNode{
			"_||_", []Node{
				&OperatorNode{"_>_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}},
				&OperatorNode{"_<_", []Node{&VariableNode{"c"}, &VariableNode{"d"}}},
			},
		}},
		{`a<b||c>d`, &OperatorNode{
			"_||_", []Node{
				&OperatorNode{"_<_", []Node{&VariableNode{"a"}, &VariableNode{"b"}}},
				&OperatorNode{"_>_", []Node{&VariableNode{"c"}, &VariableNode{"d"}}},
			},
		}},
		{`a < (b || c) > d`, &OperatorNode{
			"_>_", []Node{
				&OperatorNode{"_<_", []Node{
					&VariableNode{"a"},
					&OperatorNode{"_||_", []Node{&VariableNode{"b"}, &VariableNode{"c"}}},
				}},
				&VariableNode{"d"},
			},
		}},
		{`a < ((b && c) > d)`, &OperatorNode{
			"_<_", []Node{
				&VariableNode{"a"},
				&OperatorNode{"_>_", []Node{
					&OperatorNode{"_&&_", []Node{&VariableNode{"b"}, &VariableNode{"c"}}},
					&VariableNode{"d"},
				}},
			},
		}},
		{`a.b['c']`, &IndexAccessNode{
			&PropertyAccessNode{&VariableNode{"a"}, []string{"b"}},
			[]Node{&LiteralNode{types.String(`c`)}},
		}},
		{`a['b'].c`, &PropertyAccessNode{
			&IndexAccessNode{
				&VariableNode{"a"}, []Node{&LiteralNode{types.String(`b`)}},
			},
			[]string{"c"},
		}},
	}

	for _, tc := range tests {
		node, err := Parse(tc.source, true, defaultOption())

		assert.NoError(t, err)
		assert.EqualValues(t, tc.node, node, tc.source)
	}
}

func testParseError(t *testing.T) {
	// copy from grammar.testActionsParserError
	tests := []string{
		// Literal error
		`-0x123`,
		`'abc`,
		`4abc`,
		`a:b`,

		// Unknown operator
		`a + b`,
		`a = b`,
		`a & b`,
		`a ! b`,

		// Property & Index
		`a.`,
		`a.[b]`,
		`a.true`,
		`a.'b'`,
		`a.0123`,
		`a.a%v`,
		`a[.b]`,
		`123.a`,
		`123['a']`,

		// Function call
		`str.len()`,
		`(a.b)(x)`,
	}
	ste := new(syntaxError)
	for _, tc := range tests {
		_, err := Parse(tc, true, defaultOption())
		assert.Error(t, err)
		assert.ErrorAs(t, err, &ste, tc)
	}
}

func testParseOptions(t *testing.T) {
	t.Run("max-length", func(t *testing.T) {
		source := strings.Repeat("x", 101)
		opt := defaultOption()
		opt.MaxLength = 100
		_, err := recoverParse(true, source, opt)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "max length exceeded")
	})
	t.Run("max-depth", func(t *testing.T) {
		source := strings.Repeat("(", 101) + "x" + strings.Repeat(")", 101)
		opt := defaultOption()
		opt.MaxDepth = 100
		_, err := recoverParse(true, source, opt)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "max depth exceeded")
	})
	t.Run("max-error", func(t *testing.T) {
		source := strings.Repeat("str.len() && ", 10) + "str.len()"
		opt := defaultOption()
		opt.MaxError = 4
		_, err := recoverParse(true, source, opt)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "more than 5 errors occurred")
	})
}

func recoverParse(pureExpr bool, source string, opt Option) (node Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			ex, ok := r.(error)
			if !ok {
				ex = fmt.Errorf("panic: %v", r)
			}
			err = ex
		}
	}()

	return Parse(source, pureExpr, opt)
}

func TestParseTemplate(t *testing.T) {
	t.Run("raw-string", testParseTemplateRawString)
	t.Run("single-expr", testParseTemplateSingleExpr)
	t.Run("with-expr", testParseTemplateWithExpr)
	t.Run("error", testParseTemplateError)
}

func testParseTemplateRawString(t *testing.T) {
	tests := []string{
		`foobar`,
		`$`,
		`${`,
		`${xxxxx`,
		`${x${xxx`,
		`${we}${❤️}${δράση}`,
		`$x{xxxx`,
		`$x{{xxx`,
		`}}`,
		`}}}}}}`,
		`'foobar'`,
		`"foobar"`,
	}
	for _, input := range tests {
		node, err := Parse(input, false, defaultOption())
		expected := &LiteralNode{types.String(formatEscaper.Replace(input))}

		assert.NoError(t, err)
		assert.EqualValues(t, expected, node, input)
	}
}

func testParseTemplateSingleExpr(t *testing.T) {
	tests := map[string]Node{
		`${{ a }}`:         &VariableNode{"a"},
		`${{ null }}`:      &LiteralNode{types.NULL},
		`${{ true }}`:      &LiteralNode{types.TRUE},
		`${{ false }}`:     &LiteralNode{types.FALSE},
		`${{ Infinity }}`:  &LiteralNode{types.POSITIVE_INF},
		`${{ -Infinity }}`: &LiteralNode{types.NEGATIVE_INF},
		`${{ 'foobar' }}`:  &LiteralNode{types.String("foobar")},
		`${{ '${{' }}`:     &LiteralNode{types.String("${{")},
	}

	for source, expected := range tests {
		node, err := Parse(source, false, defaultOption())

		assert.NoError(t, err)
		assert.EqualValues(t, expected, node, source)
	}
}

func testParseTemplateWithExpr(t *testing.T) {
	tests := map[string]Node{
		`abc${{ a }}xyz`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("abc{0}xyz")},
			&VariableNode{"a"},
		}},
		`a{0}bc${{ a }}xyz`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("a{{0}}bc{0}xyz")},
			&VariableNode{"a"},
		}},
		`a{bc${{ a }}xy}z`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("a{{bc{0}xy}}z")},
			&VariableNode{"a"},
		}},
		`abc${{ '${{' }}xyz`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("abc{0}xyz")},
			&LiteralNode{types.String("${{")},
		}},
		`$${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${0}")},
			&LiteralNode{types.String("${{")},
		}},
		`${${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{{0}")},
			&LiteralNode{types.String("${{")},
		}},
		`${}${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{}}{0}")},
			&LiteralNode{types.String("${{")},
		}},
		`${x}${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{x}}{0}")},
			&LiteralNode{types.String("${{")},
		}},
		`${{ a }}${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("{0}{1}")},
			&VariableNode{"a"},
			&LiteralNode{types.String("${{")},
		}},
		`${{ a }}${x}${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("{0}${{x}}{1}")},
			&VariableNode{"a"},
			&LiteralNode{types.String("${{")},
		}},
		`${xx${yy${{ '${{' }}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{xx${{yy{0}")},
			&LiteralNode{types.String("${{")},
		}},
		"${xx\n${yy${{ '${{' }}": &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{xx\n${{yy{0}")},
			&LiteralNode{types.String("${{")},
		}},
		`${we}${{ '❤️' }}${δράση}`: &FunctionNode{"format", []Node{
			&LiteralNode{types.String("${{we}}{0}${{δράση}}")},
			&LiteralNode{types.String("❤️")},
		}},
	}

	for source, expected := range tests {
		node, err := Parse(source, false, defaultOption())

		assert.NoError(t, err)
		assert.EqualValues(t, expected, node, source)
	}
}

func testParseTemplateError(t *testing.T) {
	tests := []string{
		`${{ a`,
		`${{ a }`,
		`${{ a && }}`,
		`${{ a ${{ b }} }}`,
		`${{ a }}-${{ b }`,
	}

	ste := new(syntaxError)
	for _, tc := range tests {
		_, err := Parse(tc, false, defaultOption())
		assert.Error(t, err)
		assert.ErrorAs(t, err, &ste, tc)

		pretc := "pre" + tc
		_, err = Parse(pretc, false, defaultOption())
		assert.Error(t, err)
		assert.ErrorAs(t, err, &ste, pretc)

		subtc := tc + "sub"
		_, err = Parse(subtc, false, defaultOption())
		assert.Error(t, err)
		assert.ErrorAs(t, err, &ste, subtc)
	}
}
