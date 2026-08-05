/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unmarshal[M any](in string, fn func(M, *testing.T), opts ...jsontext.Options) func(t *testing.T) {
	return func(t *testing.T) {
		var got M
		err := json.Unmarshal([]byte(in), &got, opts...)
		require.NoError(t, err)

		fn(got, t)
	}
}

func TestContainerSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(Container, *testing.T)
	}{
		"string shorthand": {
			input: `"node:22"`,
			fn: func(got Container, t *testing.T) {
				assert.Equal(t, "node:22", got.Image)
			},
		},
		"full-form": {
			input: `{"image":"node:22","env":{"NODE_ENV":"test"},"ports":["3000"]}`,
			fn: func(got Container, t *testing.T) {
				assert.Equal(t, "node:22", got.Image)
				assert.EqualValues(t, map[string]string{"NODE_ENV": "test"}, got.Env)
				assert.EqualValues(t, []string{"3000"}, got.Ports)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got Container
		err := json.Unmarshal([]byte(`true`), &got)
		assert.ErrorContains(t, err, "expected string or object for Container")
	})
}
