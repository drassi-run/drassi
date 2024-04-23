package parser

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/operators"
	"github.com/dungdm93/drasi/pkg/expression/shared"
)

func newNodeFromToken(tk *lexicalToken) interfaces.IExpressionNode {
	switch tk.Kind() {
	case lexicalTokenKindStartIndex, lexicalTokenKindDereference:
		return new(index)
	case lexicalTokenKindLogicalOperator:
		switch tk.RawValue() {
		case shared.Not:
			return new(operators.Not)
		case shared.NotEqual:
			return new(operators.NotEqual)
		case shared.GreaterThan:
			return new(operators.GreaterThan)
		case shared.GreaterThanOrEqual:
			return new(operators.GreaterThanOrEqual)
		case shared.LessThan:
			return new(operators.LessThan)
		case shared.LessThanOrEqual:
			return new(operators.LessThanOrEqual)
		case shared.Equal:
			return new(operators.Equal)
		case shared.And:
			return new(operators.And)
		case shared.Or:
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
