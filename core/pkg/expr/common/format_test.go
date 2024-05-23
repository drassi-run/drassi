package common

import (
	"math"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/dungdm93/drassi/core/pkg/expr"
)

// TestParseNumber tests various scenarios for the ParseFloat function
func Test_ParseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		isNaN    bool // indicates whether the expected result is NaN
	}{
		{"Empty String", "", 0, false},
		{"Whitespace String", "   ", 0, false},
		{"Valid Number", "42", 42, false},
		{"Valid Float", "3.14", 3.14, false},
		{"Invalid Number", "not_a_number", math.NaN(), true},
		{"Hexadecimal", "0x2A", 42, false},
		{"Incomplete Hexadecimal", "0xz4", math.NaN(), true},
		{"Infinity", Infinity, math.Inf(1), false},
		{"Negative Infinity", NegativeInfinity, math.Inf(-1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFloat(tt.input)
			if tt.isNaN {
				if !math.IsNaN(got) {
					t.Errorf("ParseFloat(%q) = %v, want NaN", tt.input, got)
				}
			} else {
				if got != tt.expected {
					t.Errorf("ParseFloat(%q) = %v, want %v", tt.input, got, tt.expected)
				}
			}
		})
	}
}

// TestEscapeSingleQuotes tests the escapeSingleQuotes function
func Test_EscapeSingleQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty String", "", ""},
		{"No Quotes", "hello world", "hello world"},
		{"Single Quote", "it's a test", "it''s a test"},
		{"Multiple Quotes", "test 'quoted' string", "test ''quoted'' string"},
		{"Start And End Quotes", "'quoted'", "''quoted''"},
		{"Only Quotes", "'", "''"},
		{"Consecutive Quotes", "''", "''''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeSingleQuotes(tt.input)
			if got != tt.expected {
				t.Errorf("escapeSingleQuotes(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func Test_ToCanonicalValue(t *testing.T) {
	testCases := []struct {
		desc     string
		input    any
		expected any
		kind     expr.ResultKind
	}{
		{
			desc:     "Nil input",
			input:    nil,
			expected: nil,
			kind:     expr.Null,
		},
		{
			desc:     "Bool input (true)",
			input:    true,
			expected: true,
			kind:     expr.Boolean,
		},
		{
			desc:     "Bool input (false)",
			input:    false,
			expected: false,
			kind:     expr.Boolean,
		},
		{
			desc:     "Float64 input",
			input:    3.14,
			expected: 3.14,
			kind:     expr.Number,
		},
		{
			desc:     "Int input",
			input:    42,
			expected: 42,
			kind:     expr.Number,
		},
		{
			desc:     "String input",
			input:    "test",
			expected: "test",
			kind:     expr.String,
		},
		{
			desc:     "Obj input",
			input:    Obj{},
			expected: Obj{},
			kind:     expr.Object,
		},
		{
			desc:     "Array input",
			input:    Array{1, 2, 3},
			expected: Array{1, 2, 3},
			kind:     expr.Array,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			actualValue, actualKind := ToCanonicalValue(tc.input)
			assert.DeepEqual(t, tc.expected, actualValue)
			assert.Equal(t, tc.kind, actualKind)
		})
	}
}
