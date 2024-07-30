package ast

import (
	"drassi.run/core/pkg/expression/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple", testParseSimple)
	t.Run("ops-order", testParseOperatorOrder)
	t.Run("error", testParseError)
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
		node, err := Parse(tc.source)

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
		node, err := Parse(tc.source)

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
		_, err := Parse(tc)
		assert.Error(t, err)
		assert.ErrorAs(t, err, &ste, tc)
	}
}
