package scanner

import (
	"math"
	"testing"

	"github.com/dungdm93/drassi/core/pkg/expr/token"
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
		expected []token.Kind
	}{
		{
			name:     "Empty expr",
			expr:     "",
			expected: []token.Kind{},
		},
		{
			name:     "Complex expr",
			expr:     "true && (10 > 5 || 'hello' == 'world')",
			expected: []token.Kind{token.Boolean, token.LogicalOperator, token.StartGroup, token.Number, token.LogicalOperator, token.Number, token.LogicalOperator, token.Str, token.LogicalOperator, token.Str, token.EndGroup},
		},
		{
			name:     "Unclosed tokens",
			expr:     "(2 != 3",
			expected: []token.Kind{token.StartGroup, token.Number, token.LogicalOperator, token.Number},
		},
		{
			name:     "Access property name",
			expr:     "github.actor",
			expected: []token.Kind{token.NamedValue, token.Dereference, token.PropertyName},
		},
		{
			name:     "String literal",
			expr:     "'It''s open source!'",
			expected: []token.Kind{token.Str},
		},
		{
			name:     "Compare",
			expr:     "( 2 > 3 )",
			expected: []token.Kind{token.StartGroup, token.Number, token.LogicalOperator, token.Number, token.EndGroup},
		},
		{
			name:     "Function expr",
			expr:     "contains('Hello world', 'llo')",
			expected: []token.Kind{token.Function, token.StartParameters, token.Str, token.Separator, token.Str, token.EndParameters},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewScanner(tc.expr)
			var tokens []*token.Token

			for {
				token, haveResult := lexer.Next()
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
				if tokens[i].Kind() != expected {
					t.Errorf("unexpected token k at pos %d: expected %v, got %v, value %v", i, expected,
						tokens[i].Kind(), tokens[i].RawVal)
				}
			}
		})
	}
}

func TestUnclosedTokens(t *testing.T) {
	testCases := []struct {
		name             string
		expr             string
		expectedUnclosed []token.Kind
	}{
		{
			name:             "No unclosed tokens",
			expr:             "1 != 2",
			expectedUnclosed: []token.Kind{},
		},
		{
			name:             "Unclosed group",
			expr:             "(1 != 2",
			expectedUnclosed: []token.Kind{token.StartGroup},
		},
		{
			name:             "Unclosed pos and group",
			expr:             "arr[0 != (2",
			expectedUnclosed: []token.Kind{token.StartIndex, token.StartGroup},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewScanner(tc.expr)
			var lastClosed []*token.Token

			for {
				token, haveResult := lexer.Next()
				if !haveResult {
					break
				}
				lastClosed = append(lastClosed, token)
			}

			unclosedTokens := lexer.GetUnclosedTokens()

			if len(unclosedTokens) != len(tc.expectedUnclosed) {
				t.Errorf("Expected %d unclosed tokens, but got %d", len(tc.expectedUnclosed), len(unclosedTokens))
				return
			}

			for i, expected := range tc.expectedUnclosed {
				if unclosedTokens[i].Kind() != expected {
					t.Errorf("unexpected unclosed token at pos %d: expected %v, got %v", i, expected, unclosedTokens[i])
				}
			}
		})
	}
}

func TestCreateToken(t *testing.T) {
	testCases := []struct {
		name        string
		kind        token.Kind
		rawValue    string
		startIndex  int
		parsedValue any
		expected    *token.Token
	}{
		{
			name:        "Valid number token",
			kind:        token.Number,
			rawValue:    "42.0",
			startIndex:  0,
			parsedValue: 42.0,
			expected: &token.Token{
				K:         token.Number,
				RawVal:    "42.0",
				Pos:       0,
				ParsedVal: 42.0,
			},
		},
		{
			name:        "Valid string token",
			kind:        token.Str,
			rawValue:    "'hello'",
			startIndex:  0,
			parsedValue: "hello",
			expected: &token.Token{
				K:         token.Str,
				RawVal:    "'hello'",
				Pos:       0,
				ParsedVal: "hello",
			},
		},
		{
			name:        "unexpected token",
			kind:        token.Unexpected,
			rawValue:    "!@#",
			startIndex:  0,
			parsedValue: nil,
			expected: &token.Token{
				K:      token.Unexpected,
				RawVal: "!@#",
				Pos:    0,
			},
		},
	}

	for _, tc := range testCases {
		lexer := new(Scanner)
		t.Run(tc.name, func(t *testing.T) {
			token := lexer.newToken(tc.kind, tc.rawValue, tc.startIndex, tc.parsedValue)
			if token.Kind() != tc.expected.Kind() {
				t.Errorf("unexpected token K: expected %v, got %v", tc.expected.Kind(), token.Kind())
			}
			if token.RawVal != tc.expected.RawVal {
				t.Errorf("unexpected raw value: expected %v, got %v", tc.expected.RawVal, token.RawVal)
			}
			if token.Pos != tc.expected.Pos {
				t.Errorf("unexpected Pos: expected %v, got %v", tc.expected.Pos, token.Pos)
			}
			if token.ParsedVal != tc.expected.ParsedVal {
				t.Errorf("unexpected parsed value: expected %v, got %v", tc.expected.ParsedVal, token.ParsedVal)
			}
		})
	}
}

func TestReadKeywordToken(t *testing.T) {
	testCases := []struct {
		name          string
		expr          string
		expectedToken *token.Token
	}{
		{
			name: "null keyword",
			expr: "null",
			expectedToken: &token.Token{
				K:         token.Null,
				RawVal:    "null",
				Pos:       0,
				ParsedVal: nil,
			},
		},
		{
			name: "True keyword",
			expr: "true",
			expectedToken: &token.Token{
				K:         token.Boolean,
				RawVal:    "true",
				Pos:       0,
				ParsedVal: true,
			},
		},
		{
			name: "False keyword",
			expr: "false",
			expectedToken: &token.Token{
				K:         token.Boolean,
				RawVal:    "false",
				Pos:       0,
				ParsedVal: false,
			},
		},
		{
			name: "NaN keyword",
			expr: "NaN",
			expectedToken: &token.Token{
				K:         token.Number,
				RawVal:    "NaN",
				Pos:       0,
				ParsedVal: math.NaN(),
			},
		},
		{
			name: "Infinity keyword",
			expr: "Infinity",
			expectedToken: &token.Token{
				K:         token.Number,
				RawVal:    "Infinity",
				Pos:       0,
				ParsedVal: math.Inf(1),
			},
		},
		{
			name: "Fn keyword",
			expr: "someFunction()",
			expectedToken: &token.Token{
				K:         token.Function,
				RawVal:    "someFunction",
				Pos:       0,
				ParsedVal: nil,
			},
		},
		{
			name: "Named value keyword",
			expr: "someValue",
			expectedToken: &token.Token{
				K:         token.NamedValue,
				RawVal:    "someValue",
				Pos:       0,
				ParsedVal: nil,
			},
		},
		{
			name: "unexpected keyword",
			expr: "!@#$%",
			expectedToken: &token.Token{
				K:      token.Unexpected,
				RawVal: "!@#$%",
				Pos:    0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := new(Scanner)
			lexer.expr = tc.expr
			token := lexer.readKeyword()
			if token.Kind() != tc.expectedToken.Kind() {
				t.Errorf("unexpected token K: expected %v, got %v", tc.expectedToken.Kind(), token.Kind())
			}
			if token.RawVal != tc.expectedToken.RawVal {
				t.Errorf("unexpected raw value: expected %v, got %v", tc.expectedToken.RawVal, token.RawVal)
			}
			if token.Pos != tc.expectedToken.Pos {
				t.Errorf("unexpected Pos: expected %v, got %v", tc.expectedToken.Pos, token.Pos)
			}
			if diff(token.ParsedVal, tc.expectedToken.ParsedVal) {
				t.Errorf("unexpected parsed value: expected %v, got %v", tc.expectedToken.ParsedVal, token.ParsedVal)
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
