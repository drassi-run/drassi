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
		expected   []lexicalTokenKind
	}{
		{
			name:       "Empty expression",
			expression: "",
			expected:   []lexicalTokenKind{},
		},
		{
			name:       "Complex expression",
			expression: "true && (10 > 5 || 'hello' == 'world')",
			expected:   []lexicalTokenKind{lexicalTokenKindBoolean, lexicalTokenKindLogicalOperator, lexicalTokenKindStartGroup, lexicalTokenKindNumber, lexicalTokenKindLogicalOperator, lexicalTokenKindNumber, lexicalTokenKindLogicalOperator, lexicalTokenKindString, lexicalTokenKindLogicalOperator, lexicalTokenKindString, lexicalTokenKindEndGroup},
		},
		{
			name:       "Unclosed tokens",
			expression: "(2 != 3",
			expected:   []lexicalTokenKind{lexicalTokenKindStartGroup, lexicalTokenKindNumber, lexicalTokenKindLogicalOperator, lexicalTokenKindNumber},
		},
		{
			name:       "Access property name",
			expression: "github.actor",
			expected:   []lexicalTokenKind{lexicalTokenKindNamedValue, lexicalTokenKindDereference, lexicalTokenKindPropertyName},
		},
		{
			name:       "String literal",
			expression: "'It''s open source!'",
			expected:   []lexicalTokenKind{lexicalTokenKindString},
		},
		{
			name:       "Compare",
			expression: "( 2 > 3 )",
			expected:   []lexicalTokenKind{lexicalTokenKindStartGroup, lexicalTokenKindNumber, lexicalTokenKindLogicalOperator, lexicalTokenKindNumber, lexicalTokenKindEndGroup},
		},
		{
			name:       "Function expression",
			expression: "contains('Hello world', 'llo')",
			expected:   []lexicalTokenKind{lexicalTokenKindFunction, lexicalTokenKindStartParameters, lexicalTokenKindString, lexicalTokenKindSeparator, lexicalTokenKindString, lexicalTokenKindEndParameters},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer(tc.expression)
			var tokens []*lexicalToken

			for {
				token, haveResult := lexer.tryGetNextToken()
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
					t.Errorf("lexicalTokenKindUnexpected token kind at index %d: expected %v, got %s, value %v", i, expected,
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
		expectedUnclosed []lexicalTokenKind
	}{
		{
			name:             "No unclosed tokens",
			expression:       "1 != 2",
			expectedUnclosed: []lexicalTokenKind{},
		},
		{
			name:             "Unclosed group",
			expression:       "(1 != 2",
			expectedUnclosed: []lexicalTokenKind{lexicalTokenKindStartGroup},
		},
		{
			name:             "Unclosed index and group",
			expression:       "arr[0 != (2",
			expectedUnclosed: []lexicalTokenKind{lexicalTokenKindStartIndex, lexicalTokenKindStartGroup},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer(tc.expression)
			var lastClosed []*lexicalToken

			for {
				token, haveResult := lexer.tryGetNextToken()
				if !haveResult {
					break
				}
				lastClosed = append(lastClosed, token)
			}

			unclosedTokens := lexer.getUnclosedTokens()

			if len(unclosedTokens) != len(tc.expectedUnclosed) {
				t.Errorf("Expected %d unclosed tokens, but got %d", len(tc.expectedUnclosed), len(unclosedTokens))
				return
			}

			for i, expected := range tc.expectedUnclosed {
				if unclosedTokens[i].kind != expected {
					t.Errorf("lexicalTokenKindUnexpected unclosed token at index %d: expected %v, got %v", i, expected, unclosedTokens[i])
				}
			}
		})
	}
}

func TestCreateToken(t *testing.T) {
	testCases := []struct {
		name        string
		kind        lexicalTokenKind
		rawValue    string
		startIndex  int
		parsedValue any
		expected    *lexicalToken
	}{
		{
			name:        "Valid number token",
			kind:        lexicalTokenKindNumber,
			rawValue:    "42.0",
			startIndex:  0,
			parsedValue: 42.0,
			expected: &lexicalToken{
				kind:        lexicalTokenKindNumber,
				rawValue:    "42.0",
				index:       0,
				parsedValue: 42.0,
			},
		},
		{
			name:        "Valid string token",
			kind:        lexicalTokenKindString,
			rawValue:    "'hello'",
			startIndex:  0,
			parsedValue: "hello",
			expected: &lexicalToken{
				kind:        lexicalTokenKindString,
				rawValue:    "'hello'",
				index:       0,
				parsedValue: "hello",
			},
		},
		{
			name:        "lexicalTokenKindUnexpected token",
			kind:        lexicalTokenKindUnexpected,
			rawValue:    "!@#",
			startIndex:  0,
			parsedValue: nil,
			expected: &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: "!@#",
				index:    0,
			},
		},
	}

	for _, tc := range testCases {
		lexer := new(lexer)
		t.Run(tc.name, func(t *testing.T) {
			token := lexer.createToken(tc.kind, tc.rawValue, tc.startIndex, tc.parsedValue)
			if token.kind != tc.expected.kind {
				t.Errorf("lexicalTokenKindUnexpected token kind: expected %s, got %s", tc.expected.kind, token.kind)
			}
			if token.rawValue != tc.expected.rawValue {
				t.Errorf("lexicalTokenKindUnexpected raw value: expected %v, got %v", tc.expected.rawValue, token.rawValue)
			}
			if token.index != tc.expected.index {
				t.Errorf("lexicalTokenKindUnexpected index: expected %v, got %v", tc.expected.index, token.index)
			}
			if token.parsedValue != tc.expected.parsedValue {
				t.Errorf("lexicalTokenKindUnexpected parsed value: expected %v, got %v", tc.expected.parsedValue, token.parsedValue)
			}
		})
	}
}

func TestReadKeywordToken(t *testing.T) {
	testCases := []struct {
		name          string
		expression    string
		expectedToken *lexicalToken
	}{
		{
			name:       "lexicalTokenKindNull keyword",
			expression: "null",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindNull,
				rawValue:    "null",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "True keyword",
			expression: "true",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindBoolean,
				rawValue:    "true",
				index:       0,
				parsedValue: true,
			},
		},
		{
			name:       "False keyword",
			expression: "false",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindBoolean,
				rawValue:    "false",
				index:       0,
				parsedValue: false,
			},
		},
		{
			name:       "NaN keyword",
			expression: "NaN",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindNumber,
				rawValue:    "NaN",
				index:       0,
				parsedValue: math.NaN(),
			},
		},
		{
			name:       "Infinity keyword",
			expression: "Infinity",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindNumber,
				rawValue:    "Infinity",
				index:       0,
				parsedValue: math.Inf(1),
			},
		},
		{
			name:       "Fn keyword",
			expression: "someFunction()",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindFunction,
				rawValue:    "someFunction",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "Named value keyword",
			expression: "someValue",
			expectedToken: &lexicalToken{
				kind:        lexicalTokenKindNamedValue,
				rawValue:    "someValue",
				index:       0,
				parsedValue: nil,
			},
		},
		{
			name:       "lexicalTokenKindUnexpected keyword",
			expression: "!@#$%",
			expectedToken: &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: "!@#$%",
				index:    0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := new(lexer)
			lexer.expression = tc.expression
			token := lexer.readKeywordToken()
			if token.kind != tc.expectedToken.kind {
				t.Errorf("lexicalTokenKindUnexpected token kind: expected %s, got %s", tc.expectedToken.kind, token.kind)
			}
			if token.rawValue != tc.expectedToken.rawValue {
				t.Errorf("lexicalTokenKindUnexpected raw value: expected %v, got %v", tc.expectedToken.rawValue, token.rawValue)
			}
			if token.index != tc.expectedToken.index {
				t.Errorf("lexicalTokenKindUnexpected index: expected %v, got %v", tc.expectedToken.index, token.index)
			}
			if diff(token.parsedValue, tc.expectedToken.parsedValue) {
				t.Errorf("lexicalTokenKindUnexpected parsed value: expected %v, got %v", tc.expectedToken.parsedValue, token.parsedValue)
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
