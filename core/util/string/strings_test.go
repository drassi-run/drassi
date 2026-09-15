/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic ascii alphanumeric",
			input:    "Hello World 123",
			expected: "hello-world-123",
		},
		{
			name:     "preserves allowed punctuation",
			input:    "foo.bar-baz_1",
			expected: "foo.bar-baz_1",
		},
		{
			name:     "replaces special symbols with dash",
			input:    "foo/bar@baz#qux",
			expected: "foo-bar-baz-qux",
		},
		{
			name:     "Vietnamese diacritics and stroke D",
			input:    "Đặng Minh Dũng",
			expected: "dang-minh-dung",
		},
		{
			name:     "Vietnamese lowercase with multiple accents",
			input:    "tiếng việt",
			expected: "tieng-viet",
		},
		{
			name:     "German umlauts and eszett",
			input:    "München Straße Weiß",
			expected: "munchen-strasse-weiss",
		},
		{
			name:     "French accents and cedilla",
			input:    "Café naïve façade Noël",
			expected: "cafe-naive-facade-noel",
		},
		{
			name:     "Scandinavian characters",
			input:    "Tromsø Ålesund Ægir",
			expected: "tromso-alesund-aegir",
		},
		{
			name:     "Slavic crossed L",
			input:    "Kraków Łódź",
			expected: "krakow-lodz",
		},
		{
			name:     "Icelandic thorn and eth",
			input:    "Þórður",
			expected: "thordur",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Normalize(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
