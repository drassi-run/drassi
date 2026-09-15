/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package xstring

import (
	"math/rand"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const lettersLen = len(letters)

func Rand(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}

// transliterations maps Latin characters that do not decompose into ASCII base letters via Unicode NFD
// to their closest basic Latin/ASCII representations.
var transliterations = map[rune]string{
	'đ': "d",
	'Đ': "D",
	'ø': "o",
	'Ø': "O",
	'æ': "ae",
	'Æ': "AE",
	'œ': "oe",
	'Œ': "OE",
	'ß': "ss",
	'ł': "l",
	'Ł': "L",
	'þ': "th",
	'Þ': "TH",
	'ð': "d",
	'Ð': "D",
}

var diacriticTransformer = transform.Chain(
	norm.NFD,
	runes.Remove(runes.In(unicode.Mn)),
	norm.NFC,
)

func transliterate(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if sub, ok := transliterations[r]; ok {
			b.WriteString(sub)
		} else {
			b.WriteRune(r)
		}
	}

	result, _, err := transform.String(diacriticTransformer, b.String())
	if err != nil {
		return b.String()
	}
	return result
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

// Normalize
// 1. Ligature & Stroke Replacement:
// Replace Latin characters that do not decompose into base + combining mark via standard Unicode NFD:
// • Vietnamese / South Slavic: đ / Đ → d / D
// • Scandinavian: ø / Ø → o / O, æ / Æ → ae / AE, å / Å (decomposes via NFD to a)
// • German: ß → ss
// • Polish / Slavic: ł / Ł → l / L
// • Icelandic / Old English: þ / Þ → th / TH, ð / Ð → d / D
// • Latin ligatures: œ / Œ → oe / OE
// 2. Decompose & Strip Combining Diacritics:
// Use golang.org/x/text/unicode/norm and golang.org/x/text/runes:
// • Decompose with norm.NFD (e.g. é → e + \u0301, ặ → a + combining marks).
// • Strip all non-spacing marks (unicode.Mn) with runes.Remove(runes.In(unicode.Mn)).
// 3. Sanitize & Lowercase (Existing Behavior):
// • Apply nonAlphanumericRegex.ReplaceAllString(s, "-") ([^a-zA-Z0-9.\-_]+).
// • Convert to lowercase via strings.ToLower(s).
//
// NOT yet support non-Latin scripts (Cyrillic, Greek, CJK, etc.)
func Normalize(s string) string {
	s = transliterate(s)
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	s = strings.ToLower(s)
	return s
}

func EnsureSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s
	}
	return s + suffix
}

func EnsurePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}
