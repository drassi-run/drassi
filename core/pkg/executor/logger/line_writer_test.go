package logger

import (
	"github.com/stretchr/testify/assert"
	"io"
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
		n, err := io.WriteString(lw, s)
		assert.NoError(t, err)
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
	assert.EqualValues(t, expected, lines)
}
