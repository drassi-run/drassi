package parser

import (
	"fmt"
)

// lexicalTokenKind represents the kind of token in the lexer.
type lexicalTokenKind int

const (
	// lexicalTokenKindNotInitialized is used when token is nil
	lexicalTokenKindNotInitialized lexicalTokenKind = iota

	// Punctuation

	lexicalTokenKindStartGroup      // "(" logical grouping
	lexicalTokenKindStartIndex      // "["
	lexicalTokenKindStartParameters // "(" function call
	lexicalTokenKindEndGroup        // ")" logical grouping
	lexicalTokenKindEndIndex        // "]"
	lexicalTokenKindEndParameters   // ")" function call
	lexicalTokenKindSeparator       // ","
	lexicalTokenKindDereference     // "."
	lexicalTokenKindLogicalOperator // "!", "==", etc
	lexicalTokenKindWildcard        // "*"

	// Values

	lexicalTokenKindNull
	lexicalTokenKindBoolean
	lexicalTokenKindNumber
	lexicalTokenKindString
	lexicalTokenKindPropertyName
	lexicalTokenKindFunction
	lexicalTokenKindNamedValue
	lexicalTokenKindUnexpected
)

func (tk lexicalTokenKind) String() string {
	switch tk {
	case lexicalTokenKindNotInitialized:
		return "lexicalTokenKindNotInitialized"
	case lexicalTokenKindStartGroup:
		return "lexicalTokenKindStartGroup"
	case lexicalTokenKindStartIndex:
		return "lexicalTokenKindStartIndex"
	case lexicalTokenKindStartParameters:
		return "lexicalTokenKindStartParameters"
	case lexicalTokenKindEndGroup:
		return "lexicalTokenKindEndGroup"
	case lexicalTokenKindEndIndex:
		return "lexicalTokenKindEndIndex"
	case lexicalTokenKindEndParameters:
		return "lexicalTokenKindEndParameters"
	case lexicalTokenKindSeparator:
		return "lexicalTokenKindSeparator"
	case lexicalTokenKindDereference:
		return "lexicalTokenKindDereference"
	case lexicalTokenKindWildcard:
		return "lexicalTokenKindWildcard"
	case lexicalTokenKindLogicalOperator:
		return "lexicalTokenKindLogicalOperator"
	case lexicalTokenKindNull:
		return "lexicalTokenKindNull"
	case lexicalTokenKindBoolean:
		return "lexicalTokenKindBoolean"
	case lexicalTokenKindNumber:
		return "lexicalTokenKindNumber"
	case lexicalTokenKindString:
		return "lexicalTokenKindString"
	case lexicalTokenKindPropertyName:
		return "lexicalTokenKindPropertyName"
	case lexicalTokenKindFunction:
		return "Fn"
	case lexicalTokenKindNamedValue:
		return "lexicalTokenKindNamedValue"
	case lexicalTokenKindUnexpected:
		return "lexicalTokenKindUnexpected"
	default:
		return "Unknown lexicalTokenKind"
	}
}

func (tk lexicalTokenKind) ToString() string {
	return fmt.Sprintf("%s", tk)
}
