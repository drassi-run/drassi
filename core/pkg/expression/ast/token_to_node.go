package ast

import (
	"fmt"

	"drassi.run/core/pkg/expression/ast/ast_ifaces"
	"drassi.run/core/pkg/expression/ast/operators"
	"drassi.run/core/pkg/expression/common"
	"drassi.run/core/pkg/expression/token"
)

/*
ToNode converts the Token into an AST node. Different kinds of tokens result in different types of nodes:
- operators like startIndex, dereference, logicalOperator are converted to corresponding operator nodes like Index, NotEqual, GreaterThan, etc.
- literal values (null, boolean, number, str) are converted to Lit nodes containing their value.
- property names are turned into Lit nodes containing the property name as a string.
- wildcard Token is converted to a WildCard node.
Panics if it encounters an unexpected logical operator or Token kind.
*/
func ToNode(t *token.Token) ast_ifaces.ExprNode {
	switch t.K {
	case token.StartIndex, token.Dereference:
		return new(Index)
	case token.LogicalOperator:
		switch t.RawVal {
		case common.Not:
			return new(operators.Not)
		case common.NotEqual:
			return new(operators.NotEqual)
		case common.GreaterThan:
			return new(operators.GreaterThan)
		case common.GreaterThanOrEqual:
			return new(operators.GreaterThanOrEqual)
		case common.LessThan:
			return new(operators.LessThan)
		case common.LessThanOrEqual:
			return new(operators.LessThanOrEqual)
		case common.Equal:
			return new(operators.Equal)
		case common.And:
			return new(operators.And)
		case common.Or:
			return new(operators.Or)
		default:
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", t.RawVal))
		}
	case token.Null, token.Boolean, token.Number, token.Str:
		return newLiteral(t.ParsedVal)
	case token.PropertyName:
		return newLiteral(t.RawVal)
	case token.Wildcard:
		return new(WildCard)
	default:
		panic(fmt.Errorf("unexpected Token kind %v when creating node", t.Kind()))
	}
}
