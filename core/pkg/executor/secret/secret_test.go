/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package secret

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestValueSecret(t *testing.T) {
	secret := NewValueSecret("Password123!")
	t.Run("not-match", func(tt *testing.T) {
		expected := make([]*Position, 0)
		testSecret(tt, secret, "not-match-string", expected)
	})

	t.Run("simple", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
		}
		testSecret(tt, secret, "123Password123!123", expected)
	})

	t.Run("repeat", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
			{Start: 18, End: 30},
		}
		testSecret(tt, secret, "123Password123!123Password123!123", expected)
	})

	t.Run("at-start", func(tt *testing.T) {
		expected := []*Position{
			{Start: 0, End: 12},
		}
		testSecret(tt, secret, "Password123!123", expected)
	})

	t.Run("at-end", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
		}
		testSecret(tt, secret, "123Password123!", expected)
	})

	secret = NewValueSecret("foobarfoo")
	t.Run("overlap", func(tt *testing.T) {
		expected := []*Position{
			{Start: 0, End: 9},
			{Start: 6, End: 15},
			{Start: 12, End: 21},
		}
		testSecret(tt, secret, "foobarfoobarfoobarfoobar", expected)
	})
}

func TestRegexSecret(t *testing.T) {
	secret := NewRegexSecret(regexp.MustCompile("Password123!"))
	t.Run("not-match", func(tt *testing.T) {
		expected := make([]*Position, 0)
		testSecret(tt, secret, "not-match-string", expected)
	})

	t.Run("simple", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
		}
		testSecret(tt, secret, "123Password123!123", expected)
	})

	t.Run("repeat", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
			{Start: 18, End: 30},
		}
		testSecret(tt, secret, "123Password123!123Password123!123", expected)
	})

	t.Run("at-start", func(tt *testing.T) {
		expected := []*Position{
			{Start: 0, End: 12},
		}
		testSecret(tt, secret, "Password123!123", expected)
	})

	t.Run("at-end", func(tt *testing.T) {
		expected := []*Position{
			{Start: 3, End: 15},
		}
		testSecret(tt, secret, "123Password123!", expected)
	})

	secret = NewRegexSecret(regexp.MustCompile("fo{2}barfo{2}"))
	t.Run("overlap", func(tt *testing.T) {
		expected := []*Position{
			{Start: 0, End: 9},
			{Start: 6, End: 15},
			{Start: 12, End: 21},
		}
		testSecret(tt, secret, "foobarfoobarfoobarfoobar", expected)
	})

	secret = NewRegexSecret(regexp.MustCompile("a*"))
	t.Run("wildcard", func(tt *testing.T) {
		expected := []*Position{
			{Start: 0, End: 5},
			{Start: 1, End: 5},
			{Start: 2, End: 5},
			{Start: 3, End: 5},
			{Start: 4, End: 5},
		}
		testSecret(tt, secret, "aaaaa", expected)
	})
}

func testSecret(tt *testing.T, secret Secret, input string, expected []*Position) {
	actual := secret.At(input)
	assert.EqualValues(tt, expected, actual)
}
