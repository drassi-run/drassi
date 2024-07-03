package executor

import (
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestCommandController_ParseEnvFile(t *testing.T) {
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
}

func testParseEnvFile(tt *testing.T, data string, expected map[string]string) {
	reader := strings.NewReader(data)
	cc := &commandController{}
	actual, err := cc.parseEnvVars(reader)
	assert.NilError(tt, err)
	assert.DeepEqual(tt, expected, actual)
}
