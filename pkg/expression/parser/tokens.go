package parser

import (
	"github.com/dungdm93/drasi/pkg/expression/constants"
)

// LexicalToken is only usable when created with Lexer
type LexicalToken struct {
	kind        LexicalTokenKind
	rawValue    string
	index       int
	parsedValue any
}

// Kind return token's kind. If a token is created manually with default zero value for all fields, the default value of LTKNotInitialized will be returned when calling t.kind()
func (t *LexicalToken) Kind() LexicalTokenKind {
	if t == nil {
		return LTKNotInitialized
	}
	return t.kind
}

func (t *LexicalToken) RawValue() string {
	return t.rawValue
}

func (t *LexicalToken) ParsedValue() any {
	return t.parsedValue
}

func (t *LexicalToken) Index() int {
	return t.index
}

func (t *LexicalToken) Associativity() Associativity {
	if t.kind == LTKStartGroup {
		return AssociativityNone
	}
	if t.kind == LTKLogicalOperator && t.rawValue == constants.Not {
		return AssociativityRTL
	}
	if t.IsOperator() {
		return AssociativityLTR
	}
	return AssociativityRTL
}

func (t *LexicalToken) IsOperator() bool {
	switch t.kind {
	case LTKStartGroup, LTKStartIndex, LTKStartParameters, LTKEndGroup, LTKEndIndex,
		LTKEndParameters, LTKSeparator, LTKDereference, LTKLogicalOperator:
		return true
	default:
		return false
	}
}

// Precedence represents Operator precedence. The value is only meaningful for operator tokens.
func (t *LexicalToken) Precedence() int {
	switch t.kind {
	case LTKStartGroup:
		return 20
	case LTKStartIndex, LTKStartParameters, LTKDereference:
		return 19
	case LTKLogicalOperator:
		switch t.rawValue {
		case constants.Not:
			return 16
		case constants.GreaterThan, constants.GreaterThanOrEqual, constants.LessThan, constants.LessThanOrEqual:
			return 11
		case constants.NotEqual, constants.Equal:
			return 10
		case constants.And:
			return 6
		case constants.Or:
			return 5
		}
		break
	case LTKEndGroup, LTKEndIndex, LTKEndParameters, LTKSeparator:
		return 1
	}
	return 0
}

func (t *LexicalToken) OperandCount() int {
	switch t.kind {
	case LTKStartIndex, LTKDereference:
		return 2
	case LTKLogicalOperator:
		switch t.rawValue {
		case constants.Not:
			return 1
		case constants.GreaterThan, constants.GreaterThanOrEqual, constants.LessThan, constants.LessThanOrEqual, constants.Equal, constants.NotEqual, constants.And, constants.Or:
			return 2
		}
	}
	return 0
}

func IsLegalKeyWord(str string) bool {
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
