package parser

import (
	"math"
	"testing"
)

func TestTokenBoundary(t *testing.T) {
	tests := []struct {
		input    rune
		expected bool
	}{
		// Test cases for special characters
		{'(', true},
		{'[', true},
		{')', true},
		{']', true},
		{',', true},
		{'.', true},
		{'!', true},
		{'>', true},
		{'<', true},
		{'=', true},
		{'&', true},
		{'|', true},
		// Test cases for whitespace characters
		{' ', true},
		{'\t', true},
		{'\n', true},
		{'\r', true},
		// Test cases for non-special characters
		{'a', false},
		{'1', false},
		{'/', false},
		{'_', false},
		{'+', false},
	}

	for _, test := range tests {
		result := testTokenBoundary(test.input)
		if result != test.expected {
			t.Errorf("For input %q, expected %t but got %t", test.input, test.expected, result)
		}
	}
}

func TestTryGetNextToken(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
		expected   []LexicalTokenKind
	}{
		{
			name:       "Empty expression",
			expression: "",
			expected:   []LexicalTokenKind{},
		},
		{
			name:       "Complex expression",
			expression: "true && (10 > 5 || 'hello' == 'world')",
			expected:   []LexicalTokenKind{LTKBoolean, LTKLogicalOperator, LTKStartGroup, LTKNumber, LTKLogicalOperator, LTKNumber, LTKLogicalOperator, LTKString, LTKLogicalOperator, LTKString, LTKEndGroup},
		},
		{
			name:       "Unclosed tokens",
			expression: "(2 != 3",
			expected:   []LexicalTokenKind{LTKStartGroup, LTKNumber, LTKLogicalOperator, LTKNumber},
		},
		{
			name:       "Access property name",
			expression: "github.actor",
			expected:   []LexicalTokenKind{LTKNamedValue, LTKDereference, LTKPropertyName},
		},
		{
			name:       "String literal",
			expression: "'It''s open source!'",
			expected:   []LexicalTokenKind{LTKString},
		},
		{
			name:       "Compare",
			expression: "( 2 > 3 )",
			expected:   []LexicalTokenKind{LTKStartGroup, LTKNumber, LTKLogicalOperator, LTKNumber, LTKEndGroup},
		},
		{
			name:       "Function expression",
			expression: "contains('Hello world', 'llo')",
			expected:   []LexicalTokenKind{LTKFunction, LTKStartParameters, LTKString, LTKSeparator, LTKString, LTKEndParameters},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.expression)
			var tokens []*LexicalToken

			for {
				token, haveResult := lexer.TryGetNextToken()
				if !haveResult {
					break
				}
				tokens = append(tokens, token)
			}

			if len(tokens) != len(tc.expected) {
				t.Errorf("Expected %d tokens, but got %d", len(tc.expected), len(tokens))
				return
			}

			for i, expected := range tc.expected {
				if tokens[i].kind != expected {
					t.Errorf("LTKUnexpected token kind at index %d: expected %v, got %s, value %v", i, expected,
						tokens[i].kind, tokens[i].rawValue)
				}
			}
		})
	}
}

func TestUnclosedTokens(t *testing.T) {
	testCases := []struct {
		name             string
		expression       string
		expectedUnclosed []LexicalTokenKind
	}{
		{
			name:             "No unclosed tokens",
			expression:       "1 != 2",
			expectedUnclosed: []LexicalTokenKind{},
		},
		{
			name:             "Unclosed group",
			expression:       "(1 != 2",
			expectedUnclosed: []LexicalTokenKind{LTKStartGroup},
		},
		{
			name:             "Unclosed index and group",
			expression:       "arr[0 != (2",
			expectedUnclosed: []LexicalTokenKind{LTKStartIndex, LTKStartGroup},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewLexer(tc.expression)
			var lastClosed []*LexicalToken

			for {
				token, haveResult := lexer.TryGetNextToken()
				if !haveResult {
					break
				}
				lastClosed = append(lastClosed, token)
			}

			unclosedTokens := lexer.UnclosedTokens()

			if len(unclosedTokens) != len(tc.expectedUnclosed) {
				t.Errorf("Expected %d unclosed tokens, but got %d", len(tc.expectedUnclosed), len(unclosedTokens))
				return
			}

			for i, expected := range tc.expectedUnclosed {
				if unclosedTokens[i].kind != expected {
					t.Errorf("LTKUnexpected unclosed token at index %d: expected %v, got %v", i, expected, unclosedTokens[i])
				}
			}
		})
	}
}

func TestCreateToken(t *testing.T) {
	testCases := []struct {
		name        string
		kind        LexicalTokenKind
		rawValue    string
		startIndex  int
		parsedValue any
		expected    *LexicalToken
	}{
		{
			name:        "Valid number token",
			kind:        LTKNumber,
			rawValue:    "42.0",
			startIndex:  0,
			parsedValue: 42.0,
			expected: &LexicalToken{
				kind:        LTKNumber,
				rawValue:    "42.0",
				index:       0,
				parsedValue: 42.0,
			},
		},
		{
			name:        "Valid string token",
			kind:        LTKString,
			rawValue:    "'hello'",
			startIndex:  0,
			parsedValue: "hello",
			expected: &LexicalToken{
				kind:        LTKString,
				rawValue:    "'hello'",
				index:       0,
				parsedValue: "hello",
			},
		},
		{
			name:        "LTKUnexpected token",
			kind:        LTKUnexpected,
			rawValue:    "!@#",
			startIndex:  0,
			parsedValue: nil,
			expected: &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: "!@#",
				index:    0,
			},
		},
	}

	for _, tc := range testCases {
		lexer := new(Lexer)
		t.Run(tc.name, func(t *testing.T) {
			token := lexer.createToken(tc.kind, tc.rawValue, tc.startIndex, tc.parsedValue)
			if token.kind != tc.expected.kind {
				t.Errorf("LTKUnexpected token kind: expected %s, got %s", tc.expected.kind, token.kind)
			}
			if token.rawValue != tc.expected.rawValue {
				t.Errorf("LTKUnexpected raw value: expected %v, got %v", tc.expected.rawValue, token.rawValue)
			}
			if token.index != tc.expected.index {
				t.Errorf("LTKUnexpected index: expected %v, got %v", tc.expected.index, token.index)
			}
			if token.parsedValue != tc.expected.parsedValue {
				t.Errorf("LTKUnexpected parsed value: expected %v, got %v", tc.expected.parsedValue, token.parsedValue)
			}
		})
	}
}

func TestReadKeywordToken(t *testing.T) {
	testCases := []struct {
		name          string
		expression    string
		expectedToken *LexicalToken
	}{
		{
			name:       "LTKNull keyword",
			expression: "null",
			expectedToken: &LexicalToken{
				kind:        LTKNull,
				rawValue:    "null",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "True keyword",
			expression: "true",
			expectedToken: &LexicalToken{
				kind:        LTKBoolean,
				rawValue:    "true",
				index:       0,
				parsedValue: true,
			},
		},
		{
			name:       "False keyword",
			expression: "false",
			expectedToken: &LexicalToken{
				kind:        LTKBoolean,
				rawValue:    "false",
				index:       0,
				parsedValue: false,
			},
		},
		{
			name:       "NaN keyword",
			expression: "NaN",
			expectedToken: &LexicalToken{
				kind:        LTKNumber,
				rawValue:    "NaN",
				index:       0,
				parsedValue: math.NaN(),
			},
		},
		{
			name:       "Infinity keyword",
			expression: "Infinity",
			expectedToken: &LexicalToken{
				kind:        LTKNumber,
				rawValue:    "Infinity",
				index:       0,
				parsedValue: math.Inf(1),
			},
		},
		{
			name:       "Fn keyword",
			expression: "someFunction()",
			expectedToken: &LexicalToken{
				kind:        LTKFunction,
				rawValue:    "someFunction",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "Named value keyword",
			expression: "someValue",
			expectedToken: &LexicalToken{
				kind:        LTKNamedValue,
				rawValue:    "someValue",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "LTKUnexpected keyword",
			expression: "!@#$%",
			expectedToken: &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: "!@#$%",
				index:    0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := new(Lexer)
			lexer.expression = tc.expression
			token := lexer.readKeywordToken()
			if token.kind != tc.expectedToken.kind {
				t.Errorf("LTKUnexpected token kind: expected %s, got %s", tc.expectedToken.kind, token.kind)
			}
			if token.rawValue != tc.expectedToken.rawValue {
				t.Errorf("LTKUnexpected raw value: expected %v, got %v", tc.expectedToken.rawValue, token.rawValue)
			}
			if token.index != tc.expectedToken.index {
				t.Errorf("LTKUnexpected index: expected %v, got %v", tc.expectedToken.index, token.index)
			}
			if diff(token.parsedValue, tc.expectedToken.parsedValue) {
				t.Errorf("LTKUnexpected parsed value: expected %v, got %v", tc.expectedToken.parsedValue, token.parsedValue)
			}
		})
	}
}

func diff(a, b any) bool {
	if a == nil {
		return b != nil
	}
	if b == nil {
		return true
	}
	aBool, isABool := a.(bool)
	bBool, isBBool := a.(bool)
	if isABool && isBBool {
		return aBool != bBool
	}
	if math.IsNaN(a.(float64)) && math.IsNaN(b.(float64)) {
		return false
	}
	if a != b {
		return true
	}
	return false
}
