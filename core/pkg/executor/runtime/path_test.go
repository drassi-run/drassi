/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"github.com/stretchr/testify/assert"
	"iter"
	"testing"
)

func mapSeq() iter.Seq2[string, string] {
	m := [][2]string{
		{"/path/to/foo/", "/first"},
		{"/path/to/bar", "/second"},
		{"/path/to/", "/third"},
		{"/a/new/path", "/fourth"},
	}
	return func(yield func(string, string) bool) {
		for _, x := range m {
			if !yield(x[0], x[1]) {
				return
			}
		}
	}
}

func TestMapPath(t *testing.T) {
	t.Run("exact-match", func(t *testing.T) {
		r := MapPath("/path/to/foo/", mapSeq())
		assert.Equal(t, r, "/first")

		r = MapPath("/path/to/foo", mapSeq())
		assert.Equal(t, r, "/first")

		r = MapPath("/path/to/bar/", mapSeq())
		assert.Equal(t, r, "/second")

		r = MapPath("/path/to/bar", mapSeq())
		assert.Equal(t, r, "/second")
	})

	t.Run("subpath", func(t *testing.T) {
		r := MapPath("/path/to/abcxyz/", mapSeq())
		assert.Equal(t, r, "/third/abcxyz")

		r = MapPath("/path/to/abcxyz", mapSeq())
		assert.Equal(t, r, "/third/abcxyz")

		r = MapPath("/path/to/f", mapSeq()) // f is prefix of foo
		assert.Equal(t, r, "/third/f")

		r = MapPath("/path/to/foooooooooooo", mapSeq()) // foo is prefix of foooooooooooo
		assert.Equal(t, r, "/third/foooooooooooo")
	})

	t.Run("not-match", func(t *testing.T) {
		tests := []string{
			"/foobar",
			"/foo",
			"/some/random/path",
		}
		for _, o := range tests {
			r := MapPath(o, mapSeq())
			assert.Equal(t, r, "")
		}
	})
}
