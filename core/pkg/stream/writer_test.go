/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"context"
	"drassi.run/core/util/context"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestLineWriter(t *testing.T) {
	lines := make([]string, 0)
	lineHandler := HandlerFunc(func(ctx context.Context, s string) error {
		lines = append(lines, s)
		return nil
	})

	p := xcontext.NewStaticProvider(t.Context())
	lw := NewLineWriter(p, lineHandler)

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
