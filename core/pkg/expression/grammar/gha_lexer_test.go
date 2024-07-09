package grammar

import (
	"github.com/antlr4-go/antlr/v4"
	"gotest.tools/v3/assert"
	"testing"
)

type testToken struct {
	Type int
	Text string
}

func TestGHALexer(t *testing.T) {
	t.Run("null", testGHALexer("null", []testToken{{Type: GHALexerNULL, Text: "null"}}))
	t.Run("bool", testGHALexerBool)
	t.Run("integer", testGHALexerInteger)
	t.Run("float", testGHALexerFloat)
	t.Run("string", testGHALexerString)
	t.Run("identifier", testGHALexerIdentifier)
}

func testGHALexerBool(t *testing.T) {
	for _, s := range []string{"true", "false"} {
		token := testToken{Type: GHALexerBOOLEAN, Text: s}
		t.Run(s, testGHALexer(s, []testToken{token}))
	}
}

func testGHALexerInteger(t *testing.T) {
	testcases := map[string]string{
		"normal":    "1234",
		"sign+":     "+1234",
		"sign-":     "-1234",
		"leading0":  "01234",
		"leading00": "001234",
		"hex":       "0x1Afe",
		"oct":       "0o126",
	}
	for name, input := range testcases {
		token := testToken{Type: GHALexerINTEGER, Text: input}
		t.Run(name, testGHALexer(input, []testToken{token}))
	}
}

func testGHALexerString(t *testing.T) {
	testcases := map[string]string{
		"empty":        "''",
		"single-quote": "''''",
		"normal":       "'foobar'",
		"escape":       "'foo''bar'",
		"raw":          `'foo bar\tbaz\nquz'`,
	}
	for name, input := range testcases {
		token := testToken{Type: GHALexerSTRING, Text: input}
		t.Run(name, testGHALexer(input, []testToken{token}))
	}
}

func testGHALexerIdentifier(t *testing.T) {
	testcases := map[string]string{
		"single_letter": "f",
		"ignore":        "_",
		"normal":        "foobar",
		"underscore":    "_foo_bar",
		"number":        "foo123",
		"hyphen":        "foo-123",
	}
	for name, input := range testcases {
		token := testToken{Type: GHALexerIDENTIFIER, Text: input}
		t.Run(name, testGHALexer(input, []testToken{token}))
	}
}

func testGHALexerFloat(t *testing.T) {
	testcases := map[string]string{
		"inf":  "Infinity",
		"-inf": "-Infinity",
		"NaN":  "NaN",
	}
	for name, input := range testcases {
		token := testToken{Type: GHALexerFLOAT, Text: input}
		t.Run(name, testGHALexer(input, []testToken{token}))
	}

	var numbers []string
	sign := []string{"", "+", "-"}
	exponent := []string{"", "E1", "E-1", "e1", "e-1"}
	for _, s := range sign {
		for _, e := range exponent {
			if e != "" {
				numbers = append(numbers, s+"123"+e)
			}
			numbers = append(numbers, s+"123."+e, s+"123.456"+e, s+".456"+e)
		}
	}
	for _, input := range numbers {
		token := testToken{Type: GHALexerFLOAT, Text: input}
		t.Run(input, testGHALexer(input, []testToken{token}))
	}
}

func testGHALexer(input string, expected []testToken) func(t *testing.T) {
	return func(t *testing.T) {
		is := antlr.NewInputStream(input)
		lexer := NewGHALexer(is)
		cts := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		cts.Fill()
		tokens := cts.GetAllTokens()

		expected = append(expected, testToken{
			Type: antlr.TokenEOF, Text: "<EOF>",
		})

		assert.Equal(t, len(expected), len(tokens))
		for i := 0; i < len(tokens); i++ {
			actual := tokens[i]
			expect := expected[i]
			assert.Equal(t, actual.GetTokenType(), expect.Type)
			assert.Equal(t, actual.GetText(), expect.Text)
		}
	}
}
