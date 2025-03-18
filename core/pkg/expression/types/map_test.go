/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"drassi.run/core/pkg/expression/types/ref"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("val", testMapVal)
	t.Run("index", testMapIndex)
	t.Run("iterator", testMapIterator)
}

var dict = map[string]any{
	"first":  1,
	"second": 2,
	"third":  3,
}

func testMapVal(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict).(*Map),
		NewMapDynamic(dict).(*Map),
	} {
		assert.Equal(t, 3, m.Size())
		assert.Equal(t, dict, m.Value())
		assert.True(t, m.Equal(NewMapGeneric(dict)))

		for _, v := range valByType {
			assert.False(t, m.Equal(v))
		}
	}
}

func testMapIndex(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict).(*Map),
		NewMapDynamic(dict).(*Map),
	} {
		assert.Equal(t, ref.TypeString, m.IndexType())
		for k, v := range dict {
			e := m.Get(k)
			err, _ := e.(error)
			assert.NoError(t, err, "error while get key %q", k)
			assert.Equal(t, ref.TypeInteger, e.Type(), "value of key %q should be an integer", k)
			assert.EqualValues(t, v, e.Value(), "value for key %q is not equal", k)
		}
	}
}

func testMapIterator(t *testing.T) {
	for _, m := range []*Map{
		NewMapGeneric(dict).(*Map),
		NewMapDynamic(dict).(*Map),
	} {
		// track which key is accessed
		keys := make(map[string]bool, len(dict))
		for k := range dict {
			keys[k] = false
		}

		for k, v := range m.Items() {
			assert.Equal(t, ref.TypeString, k.Type(), "key %q should be a string", k)
			assert.Equal(t, ref.TypeInteger, v.Type(), "value of key %q should be an integer", k)
			nk := k.Value().(string)
			nv := dict[nk]
			assert.EqualValues(t, nv, v.Value(), "value for key %q is not equal", nk)

			keys[nk] = true
		}

		for k, v := range keys {
			assert.True(t, v, "key %q is not accessed", k) // All keys are accessed
		}
	}
}
