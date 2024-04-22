package parser

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expression/constants"
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
	"github.com/dungdm93/drasi/pkg/expression/parser/operators"
)

func newNodeFromToken(tk *LexicalToken) interfaces.IExpressionNode {
	switch tk.Kind() {
	case LTKStartIndex, LTKDereference:
		return new(Index)
	case LTKLogicalOperator:
		switch tk.RawValue() {
		case constants.Not:
			return new(operators.Not)
		case constants.NotEqual:
			return new(operators.NotEqual)
		case constants.GreaterThan:
			return new(operators.GreaterThan)
		case constants.GreaterThanOrEqual:
			return new(operators.GreaterThanOrEqual)
		case constants.LessThan:
			return new(operators.LessThan)
		case constants.LessThanOrEqual:
			return new(operators.LessThanOrEqual)
		case constants.Equal:
			return new(operators.Equal)
		case constants.And:
			return new(operators.And)
		case constants.Or:
			return new(operators.Or)
		default:
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", tk.RawValue()))
		}
	case LTKNull, LTKBoolean, LTKNumber, LTKString:
		return NewLiteral(tk.ParsedValue())
	case LTKPropertyName:
		return NewLiteral(tk.RawValue())
	case LTKWildcard:
		return new(WildCard)
	}
	panic(fmt.Errorf("unexpected kind %s when creating node", tk.Kind()))
}
