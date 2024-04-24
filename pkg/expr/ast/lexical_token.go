package ast

import (
	"fmt"

	"github.com/dungdm93/drasi/pkg/expr/ast/operators"
	"github.com/dungdm93/drasi/pkg/expr/common"
	"github.com/dungdm93/drasi/pkg/expr/interfaces"
)

// token is only usable when created with lexer
type token struct {
	k         tokenKind
	rawVal    string
	pos       int
	parsedVal any
}

// kind return token's k. If a token is created manually with default zero value for all fields, the default value of notInitialized will be returned when calling t.k()
func (t *token) kind() tokenKind {
	if t == nil {
		return notInitialized
	}
	return t.k
}

func (t *token) associativity() associativity {
	if t.k == startGroup {
		return associativityNone
	}
	if t.k == logicalOperator && t.rawVal == common.Not {
		return associativityRTL
	}
	if t.isOperator() {
		return associativityLTR
	}
	return associativityRTL
}

func (t *token) isOperator() bool {
	switch t.k {
	case startGroup, startIndex, startParameters, endGroup, endIndex,
		endParameters, separator, dereference, logicalOperator:
		return true
	default:
		return false
	}
}

// precedence represents operator precedence. The value is only meaningful for operator tokens.
func (t *token) precedence() int {
	switch t.k {
	case startGroup:
		return 20
	case startIndex, startParameters, dereference:
		return 19
	case logicalOperator:
		switch t.rawVal {
		case common.Not:
			return 16
		case common.GreaterThan, common.GreaterThanOrEqual, common.LessThan, common.LessThanOrEqual:
			return 11
		case common.NotEqual, common.Equal:
			return 10
		case common.And:
			return 6
		case common.Or:
			return 5
		}
		break
	case endGroup, endIndex, endParameters, separator:
		return 1
	}
	return 0
}

func (t *token) operandCnt() int {
	switch t.k {
	case startIndex, dereference:
		return 2
	case logicalOperator:
		switch t.rawVal {
		case common.Not:
			return 1
		case common.GreaterThan, common.GreaterThanOrEqual, common.LessThan, common.LessThanOrEqual, common.Equal, common.NotEqual, common.And, common.Or:
			return 2
		}
	}
	return 0
}

/*
toNode converts the token into an AST node. Different kinds of tokens result in different types of nodes:

operators like startIndex, dereference, logicalOperator are converted to corresponding operator nodes like index, NotEqual, GreaterThan, etc.
Literal values (null, boolean, number, str) are converted to literal nodes containing their value.
Property names are turned into literal nodes containing the property name as a string.
A wildcard token is converted to a WildCard node.
The function panics if it encounters an unexpected logical operator or token kind.
*/
func (t *token) toNode() interfaces.Node {
	switch t.k {
	case startIndex, dereference:
		return new(index)
	case logicalOperator:
		switch t.rawVal {
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
			panic(fmt.Errorf("unexpected logical operator %s when creating node ", t.rawVal))
		}
	case null, boolean, number, str:
		return newLiteral(t.parsedVal)
	case propertyName:
		return newLiteral(t.rawVal)
	case wildcard:
		return new(WildCard)
	default:
		panic(fmt.Errorf("unexpected token kind %v when creating node", t.kind()))
	}
}

// legalKeyword checks if a string qualifies as a legal keyword, which starts with an alphabet character or underscore and can contain alphabetic characters, numeric digits, underscores, or hyphens.
func legalKeyWord(str string) bool {
	if len(str) == 0 {
		return false
	}
	var first = str[0]
	if (first >= 'a' && first <= 'z') ||
		(first >= 'A' && first <= 'Z') ||
		first == '_' {
		for i := range str {
			var c = str[i]
			if (c >= 'a' && c <= 'z') ||
				(c >= 'A' && c <= 'Z') ||
				(c >= '0' && c <= '9') ||
				c == '_' ||
				c == '-' {
				// OK
			} else {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
