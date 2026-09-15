/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xstring

import (
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/transform"
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
		{
			name:     "Russian Cyrillic",
			input:    "Москва Россия Привет мир",
			expected: "moskva-rossiya-privet-mir",
		},
		{
			name:     "Ukrainian Cyrillic",
			input:    "Київ Україна",
			expected: "kiyiv-ukrayina",
		},
		{
			name:     "Greek alphabet with accents",
			input:    "Αθήνα Ελληνική Δημοκρατία",
			expected: "athena-ellenike-demokratia",
		},
		{
			name:     "Greek mixed letters and sigma variants",
			input:    "Οδυσσέας Ψυχώ",
			expected: "odysseas-psycho",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Normalize(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestTransliterator(t *testing.T) {
	t.Run("transform.String with NewTransliterator", func(t *testing.T) {
		res, _, err := transform.String(NewTransliterator(), "Київ, Αθήνα, Đặng, München")
		require.NoError(t, err)
		assert.Equal(t, "Kiyiv, Athena, Dang, Munchen", res)
	})

	t.Run("streaming with io.Reader", func(t *testing.T) {
		r := transform.NewReader(strings.NewReader("Україна & München"), NewTransliterator())
		out, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, "Ukrayina & Munchen", string(out))
	})

	t.Run("empty string", func(t *testing.T) {
		res, _, err := transform.String(NewTransliterator(), "")
		require.NoError(t, err)
		assert.Equal(t, "", res)
	})

	t.Run("concurrent usage of Normalize", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.Equal(t, "kiyiv-ukrayina", Normalize("Київ Україна"))
				assert.Equal(t, "athena-ellenike-demokratia", Normalize("Αθήνα Ελληνική Δημοκρατία"))
				assert.Equal(t, "dang-minh-dung", Normalize("Đặng Minh Dũng"))
			}()
		}
		wg.Wait()
	})
}
