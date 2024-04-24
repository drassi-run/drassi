package ast

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
		name     string
		expr     string
		expected []tokenKind
	}{
		{
			name:     "Empty expr",
			expr:     "",
			expected: []tokenKind{},
		},
		{
			name:     "Complex expr",
			expr:     "true && (10 > 5 || 'hello' == 'world')",
			expected: []tokenKind{boolean, logicalOperator, startGroup, number, logicalOperator, number, logicalOperator, str, logicalOperator, str, endGroup},
		},
		{
			name:     "Unclosed tokens",
			expr:     "(2 != 3",
			expected: []tokenKind{startGroup, number, logicalOperator, number},
		},
		{
			name:     "Access property name",
			expr:     "github.actor",
			expected: []tokenKind{namedValue, dereference, propertyName},
		},
		{
			name:     "String literal",
			expr:     "'It''s open source!'",
			expected: []tokenKind{str},
		},
		{
			name:     "Compare",
			expr:     "( 2 > 3 )",
			expected: []tokenKind{startGroup, number, logicalOperator, number, endGroup},
		},
		{
			name:     "Function expr",
			expr:     "contains('Hello world', 'llo')",
			expected: []tokenKind{function, startParameters, str, separator, str, endParameters},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer(tc.expr)
			var tokens []*token

			for {
				token, haveResult := lexer.next()
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
				if tokens[i].k != expected {
					t.Errorf("unexpected token k at pos %d: expected %v, got %v, value %v", i, expected,
						tokens[i].k, tokens[i].rawVal)
				}
			}
		})
	}
}

func TestUnclosedTokens(t *testing.T) {
	testCases := []struct {
		name             string
		expr             string
		expectedUnclosed []tokenKind
	}{
		{
			name:             "No unclosed tokens",
			expr:             "1 != 2",
			expectedUnclosed: []tokenKind{},
		},
		{
			name:             "Unclosed group",
			expr:             "(1 != 2",
			expectedUnclosed: []tokenKind{startGroup},
		},
		{
			name:             "Unclosed pos and group",
			expr:             "arr[0 != (2",
			expectedUnclosed: []tokenKind{startIndex, startGroup},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := newLexer(tc.expr)
			var lastClosed []*token

			for {
				token, haveResult := lexer.next()
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
				if unclosedTokens[i].k != expected {
					t.Errorf("unexpected unclosed token at pos %d: expected %v, got %v", i, expected, unclosedTokens[i])
				}
			}
		})
	}
}

func TestCreateToken(t *testing.T) {
	testCases := []struct {
		name        string
		kind        tokenKind
		rawValue    string
		startIndex  int
		parsedValue any
		expected    *token
	}{
		{
			name:        "Valid number token",
			kind:        number,
			rawValue:    "42.0",
			startIndex:  0,
			parsedValue: 42.0,
			expected: &token{
				k:         number,
				rawVal:    "42.0",
				pos:       0,
				parsedVal: 42.0,
			},
		},
		{
			name:        "Valid string token",
			kind:        str,
			rawValue:    "'hello'",
			startIndex:  0,
			parsedValue: "hello",
			expected: &token{
				k:         str,
				rawVal:    "'hello'",
				pos:       0,
				parsedVal: "hello",
			},
		},
		{
			name:        "unexpected token",
			kind:        unexpected,
			rawValue:    "!@#",
			startIndex:  0,
			parsedValue: nil,
			expected: &token{
				k:      unexpected,
				rawVal: "!@#",
				pos:    0,
			},
		},
	}

	for _, tc := range testCases {
		lexer := new(lexer)
		t.Run(tc.name, func(t *testing.T) {
			token := lexer.newToken(tc.kind, tc.rawValue, tc.startIndex, tc.parsedValue)
			if token.k != tc.expected.k {
				t.Errorf("unexpected token k: expected %v, got %v", tc.expected.k, token.k)
			}
			if token.rawVal != tc.expected.rawVal {
				t.Errorf("unexpected raw value: expected %v, got %v", tc.expected.rawVal, token.rawVal)
			}
			if token.pos != tc.expected.pos {
				t.Errorf("unexpected pos: expected %v, got %v", tc.expected.pos, token.pos)
			}
			if token.parsedVal != tc.expected.parsedVal {
				t.Errorf("unexpected parsed value: expected %v, got %v", tc.expected.parsedVal, token.parsedVal)
			}
		})
	}
}

func TestReadKeywordToken(t *testing.T) {
	testCases := []struct {
		name          string
		expr          string
		expectedToken *token
	}{
		{
			name: "null keyword",
			expr: "null",
			expectedToken: &token{
				k:         null,
				rawVal:    "null",
				pos:       0,
				parsedVal: nil,
			},
		},
		{
			name: "True keyword",
			expr: "true",
			expectedToken: &token{
				k:         boolean,
				rawVal:    "true",
				pos:       0,
				parsedVal: true,
			},
		},
		{
			name: "False keyword",
			expr: "false",
			expectedToken: &token{
				k:         boolean,
				rawVal:    "false",
				pos:       0,
				parsedVal: false,
			},
		},
		{
			name: "NaN keyword",
			expr: "NaN",
			expectedToken: &token{
				k:         number,
				rawVal:    "NaN",
				pos:       0,
				parsedVal: math.NaN(),
			},
		},
		{
			name: "Infinity keyword",
			expr: "Infinity",
			expectedToken: &token{
				k:         number,
				rawVal:    "Infinity",
				pos:       0,
				parsedVal: math.Inf(1),
			},
		},
		{
			name: "Fn keyword",
			expr: "someFunction()",
			expectedToken: &token{
				k:         function,
				rawVal:    "someFunction",
				pos:       0,
				parsedVal: nil,
			},
		},
		{
			name: "Named value keyword",
			expr: "someValue",
			expectedToken: &token{
				k:         namedValue,
				rawVal:    "someValue",
				pos:       0,
				parsedVal: nil,
			},
		},
		{
			name: "unexpected keyword",
			expr: "!@#$%",
			expectedToken: &token{
				k:      unexpected,
				rawVal: "!@#$%",
				pos:    0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := new(lexer)
			lexer.expr = tc.expr
			token := lexer.readKeyword()
			if token.k != tc.expectedToken.k {
				t.Errorf("unexpected token k: expected %v, got %v", tc.expectedToken.k, token.k)
			}
			if token.rawVal != tc.expectedToken.rawVal {
				t.Errorf("unexpected raw value: expected %v, got %v", tc.expectedToken.rawVal, token.rawVal)
			}
			if token.pos != tc.expectedToken.pos {
				t.Errorf("unexpected pos: expected %v, got %v", tc.expectedToken.pos, token.pos)
			}
			if diff(token.parsedVal, tc.expectedToken.parsedVal) {
				t.Errorf("unexpected parsed value: expected %v, got %v", tc.expectedToken.parsedVal, token.parsedVal)
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
