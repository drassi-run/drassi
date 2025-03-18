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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)

func Normalize(s string) string {
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
