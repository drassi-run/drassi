/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package grammar

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActionsParserExpression(t *testing.T) {
	t.Run("literal", testActionsParserLiteral)
	t.Run("identifier", testActionsParserIdentifier)
	t.Run("wrap", testActionsParserWrap)
	t.Run("logical-operator", testActionsParserLogicalOperator)
	t.Run("comparison-operator", testActionsParserComparisonOperator)
	t.Run("member", testActionsParserMember)
	t.Run("function", testActionsParserFunction)
	t.Run("error", testActionsParserError)
}

func testActionsParserLiteral(t *testing.T) {
	testcases := map[string]string{
		`null`:     `L:null`,
		`true`:     `L:true`,
		`123`:      `L:123`,
		`45.6`:     `L:45.6`,
		`'foobar'`: `L:'foobar'`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserIdentifier(t *testing.T) {
	testcases := map[string]string{
		`abc123`:     `V:abc123`,
		`_abc123`:    `V:_abc123`,
		`_abc_123`:   `V:_abc_123`,
		`_a-bc_12-3`: `V:_a-bc_12-3`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserWrap(t *testing.T) {
	testcases := map[string]string{
		`(true)`:        `(W: L:true)`,
		`(abc)`:         `(W: V:abc)`,
		`((123))`:       `(W: (W: L:123))`,
		`(a && b || c)`: `(W: (O:|| (O:&& V:a V:b) V:c))`,
		`(a || b && c)`: `(W: (O:|| V:a (O:&& V:b V:c)))`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserLogicalOperator(t *testing.T) {
	testcases := map[string]string{
		`!a`:          `(O:! V:a)`,
		`a && b`:      `(O:&& V:a V:b)`,
		`a || b`:      `(O:|| V:a V:b)`,
		`a && b && c`: `(O:&& (O:&& V:a V:b) V:c)`,
		`a || b || c`: `(O:|| (O:|| V:a V:b) V:c)`,

		// Operator Precedence
		`a && b || c`:   `(O:|| (O:&& V:a V:b) V:c)`,
		`a || b && c`:   `(O:|| V:a (O:&& V:b V:c))`,
		`(a || b) && c`: `(O:&& (W: (O:|| V:a V:b)) V:c)`,
		`a || !b`:       `(O:|| V:a (O:! V:b))`,
		`!a || b`:       `(O:|| (O:! V:a) V:b)`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserComparisonOperator(t *testing.T) {
	testcases := map[string]string{
		`a < b`:  `(O:< V:a V:b)`,
		`a <= b`: `(O:<= V:a V:b)`,
		`a > b`:  `(O:> V:a V:b)`,
		`a >= b`: `(O:>= V:a V:b)`,
		`a == b`: `(O:== V:a V:b)`,
		`a != b`: `(O:!= V:a V:b)`,

		// Operator Precedence
		`a < b == b > c`: `(O:== (O:< V:a V:b) (O:> V:b V:c))`,
		`a < !b`:         `(O:< V:a (O:! V:b))`,
		`!a > b`:         `(O:> (O:! V:a) V:b)`,
		`!a == !b`:       `(O:== (O:! V:a) (O:! V:b))`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserMember(t *testing.T) {
	testcases := map[string]string{
		// PropertyAccess
		`a.b`:   `(O:. V:a P:b)`,
		`a.b.c`: `(O:. V:a P:b P:c)`,
		`a.*`:   `(O:. V:a P:*)`,
		`a.*.c`: `(O:. V:a P:* P:c)`,
		`(a).b`: `(O:. (W: V:a) P:b)`,

		// IndexAccess
		`a['b']`:  `(O:[] V:a L:'b')`,
		`a[b]`:    `(O:[] V:a V:b)`,
		`a[b][c]`: `(O:[] V:a V:b V:c)`,
		`(a)[b]`:  `(O:[] (W: V:a) V:b)`,
		`a[(b)]`:  `(O:[] V:a (W: V:b))`,

		// Mixed
		`a.b[c]`: `(O:[] (O:. V:a P:b) V:c)`,
		`a[b].c`: `(O:. (O:[] V:a V:b) P:c)`,

		// Operator Precedence
		`!a.b`:        `(O:! (O:. V:a P:b))`,
		`!a[b]`:       `(O:! (O:[] V:a V:b))`,
		`a.b < c[d]`:  `(O:< (O:. V:a P:b) (O:[] V:c V:d))`,
		`a[b] == c.d`: `(O:== (O:[] V:a V:b) (O:. V:c P:d))`,
		`a.b && c[d]`: `(O:&& (O:. V:a P:b) (O:[] V:c V:d))`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserFunction(t *testing.T) {
	testcases := map[string]string{
		`func()`:       `(F:func)`,
		`func(a)`:      `(F:func V:a)`,
		`func('a')`:    `(F:func L:'a')`,
		`func('a', a)`: `(F:func L:'a' V:a)`,

		`func(a.b, c['d'], (true))`: `(F:func (O:. V:a P:b) (O:[] V:c L:'d') (W: L:true))`,
		`func(10 == a.b && c['d'])`: `(F:func (O:&& (O:== L:10 (O:. V:a P:b)) (O:[] V:c L:'d')))`,
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParserError(t *testing.T) {
	testcases := map[string]string{
		// Literal error
		`-0x123`: `L:-0 E:x123`, // hex is not accept sign (+/-)
		`'abc`:   ``,            // TODO: missing tail quote
		`4abc`:   `L:4 E:abc`,   // variable can't start by a number
		`a:b`:    `V:a E:b`,     // variable is not accept special char

		// Unknown operator
		`a + b`: `V:a E:b`,
		`a = b`: `V:a E:b`,
		`a & b`: `V:a E:b`,
		`a ! b`: `V:a E:! E:b`,

		// Property & Index
		`a.`:     `(O:. V:a P:<missing '*'>)`, // missing property
		`a.[b]`:  `(O:. V:a E:[ P:b) E:]`,     // redundant '.' char
		`a.true`: `(O:. V:a P:nil) E:true`,    // property MUST not be keyword
		`a.'b'`:  `(O:. V:a P:nil) E:'b'`,     // property MUST not be literal
		`a.0123`: `V:a E:.0123`,               // property MUST not valid identifier
		`a.a%v`:  `(O:. V:a P:a) E:v`,         // property MUST not contain special char
		`a[.b]`:  `(O:[] V:a E:. V:b)`,        // redundant '.' char

		`123.a`:    `L:123. E:a`,          // obj can't be literal (we can't define array/object inline)
		`123['a']`: `L:123 E:[ E:'a' E:]`, // obj can't be literal (we can't define array/object inline)

		// Function call
		`str.len()`: `(O:. V:str P:len) E:( E:)`,      // No method concept
		`(a.b)(x)`:  `(W: (O:. V:a P:b)) E:( E:x E:)`, // missing operator
	}

	for input, expected := range testcases {
		t.Run(input, testActionsParser(ActionsLexerEXPRESSION, input, expected))
	}
}

func testActionsParser(mode int, input, expected string) func(t *testing.T) {
	return func(t *testing.T) {
		is := antlr.NewInputStream(input)
		lexer := NewActionsLexer(is)
		lexer.SetMode(mode)
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parser := NewActionsParser(tokens)

		var tree antlr.ParseTree
		switch mode {
		case antlr.LexerDefaultMode:
			tree = parser.Template()
		case ActionsLexerEXPRESSION:
			tree = parser.Expression()
		default:
			err := fmt.Errorf("unknown mode: %d", mode)
			panic(err)
		}

		//// ToStringTree print out a whole tree, not just a node, in LISP format
		//// (root child1 .. childN). Print just a node if this is a leaf.
		////
		//// This show internal ruleNames e.g exprAccess. So we'd like customize it to
		//// only public nodes.
		//ruleNames := parser.GetRuleNames()
		//actual := tree.ToStringTree(ruleNames, parser)

		p := new(printListener)
		actual := p.Format(tree)

		assert.Equal(t, expected, actual)
	}
}

func TestActionsParserTemplate(t *testing.T) {
	t.Run("raw-text", testActionsParserRawText)
	t.Run("with-expr", testActionsParserWithExpr)
}

func testActionsParserRawText(t *testing.T) {
	testcases := []string{
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
	}
	for _, input := range testcases {
		expected := `T:` + input
		t.Run(input, testActionsParser(antlr.LexerDefaultMode, input, expected))
	}
}

func testActionsParserWithExpr(t *testing.T) {
	testcases := map[string]string{
		`abc${{ a }}xyz`:           `T:abc (X: V:a) T:xyz`,
		`${{ a }}`:                 `(X: V:a)`,
		`abc${{ '${{' }}xyz`:       `T:abc (X: L:'${{') T:xyz`,
		`$${{ '${{' }}`:            `T:$ (X: L:'${{')`,
		`${${{ '${{' }}`:           `T:${ (X: L:'${{')`,
		`${}${{ '${{' }}`:          `T:${} (X: L:'${{')`,
		`${x}${{ '${{' }}`:         `T:${x} (X: L:'${{')`,
		`${{ a }}${{ '${{' }}`:     `(X: V:a) (X: L:'${{')`,
		`${{ a }}${x}${{ '${{' }}`: `(X: V:a) T:${x} (X: L:'${{')`,
		`${xx${yy${{ '${{' }}`:     `T:${xx${yy (X: L:'${{')`,
		"${xx\n${yy${{ '${{' }}":   "T:${xx\n${yy (X: L:'${{')",
		`${we}${{ '❤️' }}${δράση}`: `T:${we} (X: L:'❤️') T:${δράση}`,
	}
	for input, expected := range testcases {
		t.Run(input, testActionsParser(antlr.LexerDefaultMode, input, expected))
	}
}
