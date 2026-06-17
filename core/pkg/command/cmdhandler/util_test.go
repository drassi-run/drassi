/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitLine(t *testing.T) {
	content := "foobar\n\n  whitespace-prefix\r\n     \nwhitespace-suffix    \rabcxyz"
	lines := slices.Collect(splitLine(content))
	expected := []string{"foobar", "whitespace-prefix", "whitespace-suffix", "abcxyz"}
	assert.EqualValues(t, expected, lines)
}

func TestReadLine(t *testing.T) {
	content := `
foobar

# a comment
abcxyz
`
	r := strings.NewReader(content)
	actual, err := readLine(r)
	assert.NoError(t, err)

	expected := []string{"foobar", "abcxyz"}
	assert.EqualValues(t, expected, actual)
}

func TestParseEnvVars(t *testing.T) {
	t.Run("simple", func(tt *testing.T) {
		data := "NODE_OPTIONS=asdf"
		env := map[string]string{
			"NODE_OPTIONS": "asdf",
		}
		testParseEnvFile(tt, data, env)
	})

	t.Run("heredoc/singleline", func(tt *testing.T) {
		data := `
NODE_OPTIONS<<EOF
asdf
EOF
`
		env := map[string]string{
			"NODE_OPTIONS": "asdf",
		}
		testParseEnvFile(tt, data, env)
	})

	t.Run("heredoc/multiline", func(tt *testing.T) {
		data := `
NODE_OPTIONS<<EOF
line one
line two
line three
EOF
`
		env := map[string]string{
			"NODE_OPTIONS": "line one\nline two\nline three",
		}
		testParseEnvFile(tt, data, env)
	})

	t.Run("heredoc/empty", func(tt *testing.T) {
		data := `
NODE_OPTIONS<<EOF
EOF
`
		env := map[string]string{
			"NODE_OPTIONS": "",
		}
		testParseEnvFile(tt, data, env)
	})

	t.Run("complex", func(tt *testing.T) {
		data := `
MY_ENV<<=EOF
hello
one
=EOF
MY_ENV_2<<<EOF
hello
two
<EOF
MY_ENV_3<<EOF
hello

three

EOF

ABC=xyz

MY_ENV_4<<EOF
hello=four
EOF
FOO=<<bar
MY_ENV_5<<EOF
 EOF
EOF
`
		env := map[string]string{
			"MY_ENV":   "hello\none",
			"MY_ENV_2": "hello\ntwo",
			"MY_ENV_3": "hello\n\nthree\n",
			"ABC":      "xyz",
			"MY_ENV_4": "hello=four",
			"FOO":      "<<bar",
			"MY_ENV_5": " EOF",
		}
		testParseEnvFile(tt, data, env)
	})

	t.Run("invalid", func(tt *testing.T) {
		data := []string{
			"FOOBAR",
			"=FOOBAR",
			"FOOBAR<<",
			`
<<EOF
foobar
EOF
`,
			`
FOOBAR<<EOF
foobar
`,
		}
		for _, d := range data {
			invalidParseEnvFile(tt, d)
		}
	})
}

func testParseEnvFile(tt *testing.T, data string, expected map[string]string) {
	reader := strings.NewReader(data)
	actual, err := parseEnvVars(reader)
	assert.NoError(tt, err)
	assert.EqualValues(tt, expected, actual)
}

func invalidParseEnvFile(tt *testing.T, data string) {
	reader := strings.NewReader(data)
	_, err := parseEnvVars(reader)
	assert.ErrorIs(tt, err, ErrInvalidFile)
}
