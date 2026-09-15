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
	"sync"
	"unicode"
	"unicode/utf8"

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

// transliterations maps Unicode characters to their closest basic Latin/ASCII representations.
var transliterations = map[rune]string{
	// NOTE: Latin ligatures and characters with strokes that do not decompose into ASCII via Unicode NFD.
	'đ': "d", 'Đ': "D",
	'ø': "o", 'Ø': "O",
	'æ': "ae", 'Æ': "AE",
	'œ': "oe", 'Œ': "OE",
	'ß': "ss",
	'ł': "l", 'Ł': "L",
	'þ': "th", 'Þ': "TH",
	'ð': "d", 'Ð': "D",

	// NOTE: Cyrillic characters (Russian, Ukrainian, Belarusian, Bulgarian, Serbian) using standard Romanization.
	'а': "a", 'А': "A",
	'б': "b", 'Б': "B",
	'в': "v", 'В': "V",
	'г': "g", 'Г': "G",
	'ґ': "g", 'Ґ': "G",
	'д': "d", 'Д': "D",
	'ђ': "dj", 'Ђ': "Dj",
	'е': "e", 'Е': "E",
	'ё': "yo", 'Ё': "Yo",
	'є': "ye", 'Є': "Ye",
	'ж': "zh", 'Ж': "Zh",
	'з': "z", 'З': "Z",
	'ѕ': "dz", 'Ѕ': "Dz",
	'и': "i", 'И': "I",
	'і': "i", 'І': "I",
	'ї': "yi", 'Ї': "Yi",
	'й': "y", 'Й': "Y",
	'ј': "j", 'Ј': "J",
	'к': "k", 'К': "K",
	'л': "l", 'Л': "L",
	'љ': "lj", 'Љ': "Lj",
	'м': "m", 'М': "M",
	'н': "n", 'Н': "N",
	'њ': "nj", 'Њ': "Nj",
	'о': "o", 'О': "O",
	'п': "p", 'П': "P",
	'р': "r", 'Р': "R",
	'с': "s", 'С': "S",
	'т': "t", 'Т': "T",
	'ћ': "c", 'Ћ': "C",
	'у': "u", 'У': "U",
	'ў': "u", 'Ў': "U",
	'ф': "f", 'Ф': "F",
	'х': "kh", 'Х': "Kh",
	'ц': "ts", 'Ц': "Ts",
	'ч': "ch", 'Ч': "Ch",
	'џ': "dz", 'Џ': "Dz",
	'ш': "sh", 'Ш': "Sh",
	'щ': "shch", 'Щ': "Shch",
	'ъ': "", 'Ъ': "",
	'ы': "y", 'Ы': "Y",
	'ь': "", 'Ь': "",
	'э': "e", 'Э': "E",
	'ю': "yu", 'Ю': "Yu",
	'я': "ya", 'Я': "Ya",

	// NOTE: Greek alphabet (ISO 843 Romanization; combining accents stripped via NFD).
	'α': "a", 'Α': "A",
	'β': "v", 'Β': "V",
	'γ': "g", 'Γ': "G",
	'δ': "d", 'Δ': "D",
	'ε': "e", 'Ε': "E",
	'ζ': "z", 'Ζ': "Z",
	'η': "e", 'Η': "E",
	'θ': "th", 'Θ': "Th",
	'ι': "i", 'Ι': "I",
	'κ': "k", 'Κ': "K",
	'λ': "l", 'Λ': "L",
	'μ': "m", 'Μ': "M",
	'ν': "n", 'Ν': "N",
	'ξ': "x", 'Ξ': "X",
	'ο': "o", 'Ο': "O",
	'π': "p", 'Π': "P",
	'ρ': "r", 'Ρ': "R",
	'σ': "s", 'Σ': "S", 'ς': "s",
	'τ': "t", 'Τ': "T",
	'υ': "y", 'Υ': "Y",
	'φ': "ph", 'Φ': "Ph",
	'χ': "ch", 'Χ': "Ch",
	'ψ': "ps", 'Ψ': "Ps",
	'ω': "o", 'Ω': "O",

	// NOTE: Greek vowels with tonos / dialytika (monotonic Greek).
	'ά': "a", 'Ά': "A",
	'έ': "e", 'Έ': "E",
	'ή': "e", 'Ή': "E",
	'ί': "i", 'Ί': "I", 'ϊ': "i", 'Ϊ': "I", 'ΐ': "i",
	'ό': "o", 'Ό': "O",
	'ύ': "y", 'Ύ': "Y", 'ϋ': "y", 'Ϋ': "Y", 'ΰ': "y",
	'ώ': "o", 'Ώ': "O",
}

// transliterator is a transform.Transformer that maps Unicode characters
// (Latin ligatures/strokes, Cyrillic, and Greek) to their closest basic Latin/ASCII representations.
type transliterator struct {
	transform.NopResetter
}

func (transliterator) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		r, size := utf8.DecodeRune(src[nSrc:])
		if r == utf8.RuneError && size == 1 {
			if !atEOF && !utf8.FullRune(src[nSrc:]) {
				err = transform.ErrShortSrc
				break
			}
		}

		if sub, ok := transliterations[r]; ok {
			if len(dst)-nDst < len(sub) {
				err = transform.ErrShortDst
				break
			}
			nDst += copy(dst[nDst:], sub)
			nSrc += size
		} else {
			if len(dst)-nDst < size {
				err = transform.ErrShortDst
				break
			}
			nDst += copy(dst[nDst:], src[nSrc:nSrc+size])
			nSrc += size
		}
	}
	return nDst, nSrc, err
}

var transliteratorPool = sync.Pool{
	New: func() any {
		return transform.Chain(
			new(transliterator),
			norm.NFD,
			runes.Remove(runes.In(unicode.Mn)),
			norm.NFC,
		)
	},
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
// • Also transliterates Cyrillic and Greek characters to their Latin/ASCII equivalents.
// • Note: CJK and other non-phonetic scripts are not transliterated.
func Normalize(s string) string {
	t := transliteratorPool.Get().(transform.Transformer)
	s, _, _ = transform.String(t, s)
	transliteratorPool.Put(t)
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
