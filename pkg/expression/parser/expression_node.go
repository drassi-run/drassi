package parser

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression"
	"github.com/dungdm93/drasi/pkg/expression/parser/operators"
)

func newNodeFromToken(tk *lexicalToken) expression.IExpNode {
	switch tk.Kind() {
	case lexicalTokenKindStartIndex, lexicalTokenKindDereference:
		return new(index)
	case lexicalTokenKindLogicalOperator:
		switch tk.RawValue() {
		case expression.Not:
			return new(operators.Not)
		case expression.NotEqual:
			return new(operators.NotEqual)
		case expression.GreaterThan:
			return new(operators.GreaterThan)
		case expression.GreaterThanOrEqual:
			return new(operators.GreaterThanOrEqual)
		case expression.LessThan:
			return new(operators.LessThan)
		case expression.LessThanOrEqual:
			return new(operators.LessThanOrEqual)
		case expression.Equal:
			return new(operators.Equal)
		case expression.And:
			return new(operators.And)
		case expression.Or:
			return new(operators.Or)
		default:
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", tk.RawValue()))
		}
	case lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString:
		return newLiteral(tk.ParsedValue())
	case lexicalTokenKindPropertyName:
		return newLiteral(tk.RawValue())
	case lexicalTokenKindWildcard:
		return new(WildCard)
	}
	panic(fmt.Errorf("unexpected kind %s when creating node", tk.Kind()))
}
