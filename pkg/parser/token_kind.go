package parser

import (
	"fmt"
)

// LexicalTokenKind represents the kind of token in the lexer.
type LexicalTokenKind int

const (
	// LTKNotInitialized is used when token is nil
	LTKNotInitialized LexicalTokenKind = iota

	// Punctuation
	LTKStartGroup      // "(" logical grouping
	LTKStartIndex      // "["
	LTKStartParameters // "(" function call
	LTKEndGroup        // ")" logical grouping
	LTKEndIndex        // "]"
	LTKEndParameters   // ")" function call
	LTKSeparator       // ","
	LTKDereference     // "."
	LTKLogicalOperator // "!", "==", etc

	LTKWildcard // "*"

	// Values
	LTKNull
	LTKBoolean
	LTKNumber
	LTKString
	LTKPropertyName
	LTKFunction
	LTKNamedValue

	LTKUnexpected
)

func (tk LexicalTokenKind) String() string {
	switch tk {
	case LTKNotInitialized:
		return "LTKNotInitialized"
	case LTKStartGroup:
		return "LTKStartGroup"
	case LTKStartIndex:
		return "LTKStartIndex"
	case LTKStartParameters:
		return "LTKStartParameters"
	case LTKEndGroup:
		return "LTKEndGroup"
	case LTKEndIndex:
		return "LTKEndIndex"
	case LTKEndParameters:
		return "LTKEndParameters"
	case LTKSeparator:
		return "LTKSeparator"
	case LTKDereference:
		return "LTKDereference"
	case LTKWildcard:
		return "LTKWildcard"
	case LTKLogicalOperator:
		return "LTKLogicalOperator"
	case LTKNull:
		return "LTKNull"
	case LTKBoolean:
		return "LTKBoolean"
	case LTKNumber:
		return "LTKNumber"
	case LTKString:
		return "LTKString"
	case LTKPropertyName:
		return "LTKPropertyName"
	case LTKFunction:
		return "Fn"
	case LTKNamedValue:
		return "LTKNamedValue"
	case LTKUnexpected:
		return "LTKUnexpected"
	default:
		return "Unknown LexicalTokenKind"
	}
}

func (tk LexicalTokenKind) ToString() string {
	return fmt.Sprintf("%s", tk)
}
