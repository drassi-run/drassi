/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package grammar

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testToken struct {
	Type int
	Text string
}

func TestActionsLexerExpression(t *testing.T) {
	t.Run("null", testActionsLexer(ActionsLexerEXPRESSION, "null", []testToken{{Type: ActionsLexerNULL, Text: "null"}}))
	t.Run("bool", testActionsLexerBool)
	t.Run("integer", testActionsLexerInteger)
	t.Run("float", testActionsLexerFloat)
	t.Run("string", testActionsLexerString)
	t.Run("identifier", testActionsLexerIdentifier)
}

func testActionsLexerBool(t *testing.T) {
	for _, s := range []string{"true", "false"} {
		token := testToken{Type: ActionsLexerBOOLEAN, Text: s}
		t.Run(s, testActionsLexer(ActionsLexerEXPRESSION, s, []testToken{token}))
	}
}

func testActionsLexerInteger(t *testing.T) {
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
		token := testToken{Type: ActionsLexerINTEGER, Text: input}
		t.Run(name, testActionsLexer(ActionsLexerEXPRESSION, input, []testToken{token}))
	}
}

func testActionsLexerString(t *testing.T) {
	testcases := map[string]string{
		"empty":        "''",
		"single-quote": "''''",
		"normal":       "'foobar'",
		"escape":       "'foo''bar'",
		"raw":          `'foo bar\tbaz\nquz'`,
	}
	for name, input := range testcases {
		token := testToken{Type: ActionsLexerSTRING, Text: input}
		t.Run(name, testActionsLexer(ActionsLexerEXPRESSION, input, []testToken{token}))
	}
}

func testActionsLexerIdentifier(t *testing.T) {
	testcases := map[string]string{
		"single_letter": "f",
		"ignore":        "_",
		"normal":        "foobar",
		"underscore":    "_foo_bar",
		"number":        "foo123",
		"hyphen":        "foo-123",
	}
	for name, input := range testcases {
		token := testToken{Type: ActionsLexerIDENTIFIER, Text: input}
		t.Run(name, testActionsLexer(ActionsLexerEXPRESSION, input, []testToken{token}))
	}
}

func testActionsLexerFloat(t *testing.T) {
	testcases := map[string]string{
		"inf":  "Infinity",
		"-inf": "-Infinity",
		"NaN":  "NaN",
	}
	for name, input := range testcases {
		token := testToken{Type: ActionsLexerFLOAT, Text: input}
		t.Run(name, testActionsLexer(ActionsLexerEXPRESSION, input, []testToken{token}))
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
		token := testToken{Type: ActionsLexerFLOAT, Text: input}
		t.Run(input, testActionsLexer(ActionsLexerEXPRESSION, input, []testToken{token}))
	}
}

func testActionsLexer(mode int, input string, expected []testToken) func(t *testing.T) {
	return func(t *testing.T) {
		is := antlr.NewInputStream(input)
		lexer := NewActionsLexer(is)
		lexer.SetMode(mode)
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

func TestActionsLexerTemplate(t *testing.T) {
	t.Run("raw-text", testActionsLexerRawText)
	t.Run("with-expr", testActionsLexerWithExpr)
}

func testActionsLexerRawText(t *testing.T) {
	tests := map[string][]string{
		`foobar`:             {`foobar`},
		`$`:                  {`$`},
		`${`:                 {`${`},
		`${xxxxx`:            {`${xxxxx`},
		`${x${xxx`:           {`${x`, `${xxx`},
		`${we}${❤️}${δράση}`: {`${we}`, `${❤️}`, `${δράση}`},
		`$x{xxxx`:            {`$x{xxxx`},
		`$x{{xxx`:            {`$x{{xxx`},
		`}}`:                 {`}}`},
		`}}}}}}`:             {`}}}}}}`},
	}

	for input, exp := range tests {
		expected := make([]testToken, len(exp))
		for i, e := range exp {
			expected[i] = testToken{Type: ActionsLexerTEXT, Text: e}
		}
		t.Run(input, testActionsLexer(antlr.LexerDefaultMode, input, expected))
	}
}

func testActionsLexerWithExpr(t *testing.T) {
	tests := map[string][]testToken{
		`abc${{ a }}xyz`: {
			{ActionsLexerTEXT, `abc`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerIDENTIFIER, `a`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
			{ActionsLexerTEXT, `xyz`},
		},
		`${{ a }}`: {
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerIDENTIFIER, `a`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`abc${{ '${{' }}xyz`: {
			{ActionsLexerTEXT, `abc`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
			{ActionsLexerTEXT, `xyz`},
		},
		`$${{ '${{' }}`: {
			{ActionsLexerTEXT, `$`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${${{ '${{' }}`: {
			{ActionsLexerTEXT, `${`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${}${{ '${{' }}`: {
			{ActionsLexerTEXT, `${}`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${x}${{ '${{' }}`: {
			{ActionsLexerTEXT, `${x}`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${{ a }}${{ '${{' }}`: {
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerIDENTIFIER, `a`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${{ a }}${x}${{ '${{' }}`: {
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerIDENTIFIER, `a`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
			{ActionsLexerTEXT, `${x}`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${xx${yy${{ '${{' }}`: {
			{ActionsLexerTEXT, `${xx`}, {ActionsLexerTEXT, `${yy`}, // NOTE `${xx${yy` is split into 2 tokens as current implementation
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		"${xx\n${yy${{ '${{' }}": {
			{ActionsLexerTEXT, "${xx\n"}, {ActionsLexerTEXT, `${yy`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'${{'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
		},
		`${we}${{ '❤️' }}${δράση}`: {
			{ActionsLexerTEXT, `${we}`},
			{ActionsLexerEXPRESSION_OPEN, `${{`}, {ActionsLexerSTRING, `'❤️'`}, {ActionsLexerEXPRESSION_CLOSE, `}}`},
			{ActionsLexerTEXT, `${δράση}`},
		},
	}

	for input, expected := range tests {
		t.Run(input, testActionsLexer(antlr.LexerDefaultMode, input, expected))
	}
}
