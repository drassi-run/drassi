package parser

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

// lexicalToken is only usable when created with lexer
type lexicalToken struct {
	kind        lexicalTokenKind
	rawValue    string
	index       int
	parsedValue any
}

// Kind return token's kind. If a token is created manually with default zero value for all fields, the default value of lexicalTokenKindNotInitialized will be returned when calling t.kind()
func (t *lexicalToken) Kind() lexicalTokenKind {
	if t == nil {
		return lexicalTokenKindNotInitialized
	}
	return t.kind
}

func (t *lexicalToken) RawValue() string {
	return t.rawValue
}

func (t *lexicalToken) ParsedValue() any {
	return t.parsedValue
}

func (t *lexicalToken) Index() int {
	return t.index
}

func (t *lexicalToken) Associativity() associativity {
	if t.kind == lexicalTokenKindStartGroup {
		return associativityNone
	}
	if t.kind == lexicalTokenKindLogicalOperator && t.rawValue == expression.Not {
		return associativityRTL
	}
	if t.IsOperator() {
		return associativityLTR
	}
	return associativityRTL
}

func (t *lexicalToken) IsOperator() bool {
	switch t.kind {
	case lexicalTokenKindStartGroup, lexicalTokenKindStartIndex, lexicalTokenKindStartParameters, lexicalTokenKindEndGroup, lexicalTokenKindEndIndex,
		lexicalTokenKindEndParameters, lexicalTokenKindSeparator, lexicalTokenKindDereference, lexicalTokenKindLogicalOperator:
		return true
	default:
		return false
	}
}

// Precedence represents Operator precedence. The value is only meaningful for operator tokens.
func (t *lexicalToken) Precedence() int {
	switch t.kind {
	case lexicalTokenKindStartGroup:
		return 20
	case lexicalTokenKindStartIndex, lexicalTokenKindStartParameters, lexicalTokenKindDereference:
		return 19
	case lexicalTokenKindLogicalOperator:
		switch t.rawValue {
		case expression.Not:
			return 16
		case expression.GreaterThan, expression.GreaterThanOrEqual, expression.LessThan, expression.LessThanOrEqual:
			return 11
		case expression.NotEqual, expression.Equal:
			return 10
		case expression.And:
			return 6
		case expression.Or:
			return 5
		}
		break
	case lexicalTokenKindEndGroup, lexicalTokenKindEndIndex, lexicalTokenKindEndParameters, lexicalTokenKindSeparator:
		return 1
	}
	return 0
}

func (t *lexicalToken) OperandCount() int {
	switch t.kind {
	case lexicalTokenKindStartIndex, lexicalTokenKindDereference:
		return 2
	case lexicalTokenKindLogicalOperator:
		switch t.rawValue {
		case expression.Not:
			return 1
		case expression.GreaterThan, expression.GreaterThanOrEqual, expression.LessThan, expression.LessThanOrEqual, expression.Equal, expression.NotEqual, expression.And, expression.Or:
			return 2
		}
	}
	return 0
}

func isLegalKeyWord(str string) bool {
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
