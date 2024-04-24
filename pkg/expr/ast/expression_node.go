package ast

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/operators"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

func tokenToNode(tk *token) interfaces.Node {
	switch tk.kind() {
	case startIndex, dereference:
		return new(index)
	case logicalOperator:
		switch tk.rawVal {
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
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", tk.rawVal))
		}
	case null, boolean, number, str:
		return newLiteral(tk.parsedVal)
	case propertyName:
		return newLiteral(tk.rawVal)
	case wildcard:
		return new(WildCard)
	default:
		panic(fmt.Errorf("unexpected token kind %v when creating node", tk.kind()))
	}
}
