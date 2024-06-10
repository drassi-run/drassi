package token

import (
	"github.com/dungdm93/drassi/core/pkg/expression/common"
)

// Kind represents the Kind of Token in the scanner.
type Kind int

const (
	// NotInitialized is used when Token is nil
	NotInitialized Kind = iota
	// punctuation
	StartGroup      // ( logical grouping
	StartIndex      // [
	StartParameters // ( function call
	EndGroup        // ) logical grouping
	EndIndex        // ]
	EndParameters   // ) function call
	Separator       // ,
	Dereference     // .
	LogicalOperator // !, ==, !=, >=, >, <, <=
	Wildcard        // *
	// values
	Null
	Boolean
	Number
	Float
	Int
	Str
	PropertyName
	Function
	NamedValue
	Unexpected
)

func (t *Token) String() string {
	return tokens[t.Kind()]
}

var tokens = [...]string{
	StartGroup:      "(",
	StartIndex:      "[",
	StartParameters: "(",
	EndGroup:        ")",
	EndIndex:        "]",
	EndParameters:   ")",
	Separator:       ",",
	Dereference:     ".",
	Wildcard:        "*",
}

// Token is only usable when created with lexer
type Token struct {
	K         Kind
	RawVal    string
	Pos       int
	ParsedVal any
}

// Kind return Token's K. If a Token is created manually with default zero value for all fields, the default value of NotInitialized will be returned when calling t.K()
func (t *Token) Kind() Kind {
	if t == nil {
		return NotInitialized
	}
	return t.K
}

func (t *Token) Associativity() Associativity {
	if t.K == StartGroup {
		return AssociativityNone
	}
	if t.K == LogicalOperator && t.RawVal == common.Not {
		return AssociativityRTL
	}
	if t.IsOperator() {
		return AssociativityLTR
	}
	return AssociativityRTL
}

func (t *Token) IsOperator() bool {
	switch t.K {
	case StartGroup, StartIndex, StartParameters, EndGroup, EndIndex,
		EndParameters, Separator, Dereference, LogicalOperator:
		return true
	default:
		return false
	}
}

// Precedence represents operator Precedence. The value is only meaningful for operator tokens.
func (op *Token) Precedence() int {
	switch op.K {
	case StartGroup:
		return 20
	case StartIndex, StartParameters, Dereference:
		return 19
	case LogicalOperator:
		switch op.RawVal {
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
	case EndGroup, EndIndex, EndParameters, Separator:
		return 1
	}
	return 0
}

func (t *Token) OperandCnt() int {
	switch t.K {
	case StartIndex, Dereference:
		return 2
	case LogicalOperator:
		switch t.RawVal {
		case common.Not:
			return 1
		case common.GreaterThan, common.GreaterThanOrEqual, common.LessThan, common.LessThanOrEqual, common.Equal, common.NotEqual, common.And, common.Or:
			return 2
		}
	}
	return 0
}

// LegalKeyWord checks if a string qualifies as a legal keyword, which starts with an alphabet character or underscore and can contain alphabetic characters, numeric digits, underscores, or hyphens.
func LegalKeyWord(str string) bool {
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
