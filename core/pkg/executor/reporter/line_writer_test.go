package reporter

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestLineWriter(t *testing.T) {
	lines := make([]string, 0)
	lineHandler := func(s string) error {
		lines = append(lines, s)
		return nil
	}

	lw := NewLineWriter(lineHandler)

	write := func(s string) {
		n, err := lw.Write([]byte(s))
		assert.NilError(t, err)
		assert.Equal(t, len(s), n, s)
	}

	write("hello")
	write(" ")
	write("world!!\nextra")
	write(" line\n and another\nlast")
	write(" line\n")
	write("no newline here...")

	expected := []string{
		"hello world!!\n",
		"extra line\n",
		" and another\n",
		"last line\n",
	}
	assert.DeepEqual(t, expected, lines)
}
