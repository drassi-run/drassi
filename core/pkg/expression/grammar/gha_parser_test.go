package grammar

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGHAParser(t *testing.T) {
	t.Run("literal", testGHAParserLiteral)
	t.Run("identifier", testGHAParserIdentifier)
	t.Run("wrap", testGHAParserWrap)
	t.Run("logical-operator", testGHAParserLogicalOperator)
	t.Run("comparison-operator", testGHAParserComparisonOperator)
	t.Run("member", testGHAParserMember)
	t.Run("function", testGHAParserFunction)
}

func testGHAParserLiteral(t *testing.T) {
	testcases := map[string]string{
		`null`:     `L:null`,
		`true`:     `L:true`,
		`123`:      `L:123`,
		`45.6`:     `L:45.6`,
		`'foobar'`: `L:'foobar'`,
	}

	for input, expected := range testcases {
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserIdentifier(t *testing.T) {
	testcases := map[string]string{
		`abc123`:     `V:abc123`,
		`_abc123`:    `V:_abc123`,
		`_abc_123`:   `V:_abc_123`,
		`_a-bc_12-3`: `V:_a-bc_12-3`,
	}

	for input, expected := range testcases {
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserWrap(t *testing.T) {
	testcases := map[string]string{
		`(true)`:        `(W: L:true)`,
		`(abc)`:         `(W: V:abc)`,
		`((123))`:       `(W: (W: L:123))`,
		`(a && b || c)`: `(W: (O:|| (O:&& V:a V:b) V:c))`,
		`(a || b && c)`: `(W: (O:|| V:a (O:&& V:b V:c)))`,
	}

	for input, expected := range testcases {
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserLogicalOperator(t *testing.T) {
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
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserComparisonOperator(t *testing.T) {
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
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserMember(t *testing.T) {
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
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParserFunction(t *testing.T) {
	testcases := map[string]string{
		`func()`:       `(F:func)`,
		`func(a)`:      `(F:func V:a)`,
		`func('a')`:    `(F:func L:'a')`,
		`func('a', a)`: `(F:func L:'a' V:a)`,

		`func(a.b, c['d'], (true))`: `(F:func (O:. V:a P:b) (O:[] V:c L:'d') (W: L:true))`,
		`func(10 == a.b && c['d'])`: `(F:func (O:&& (O:== L:10 (O:. V:a P:b)) (O:[] V:c L:'d')))`,
	}

	for input, expected := range testcases {
		t.Run(input, testGHAParser(input, expected))
	}
}

func testGHAParser(input, expected string) func(t *testing.T) {
	return func(t *testing.T) {
		is := antlr.NewInputStream(input)
		lexer := NewGHALexer(is)
		tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parser := NewGHAParser(tokens)
		tree := parser.Expression()

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
