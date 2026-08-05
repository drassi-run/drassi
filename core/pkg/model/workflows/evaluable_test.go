/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenSerde(t *testing.T) {
	opt := json.WithUnmarshalers(json.UnmarshalFromFunc(unmarshalToken))

	testcases := map[string]struct {
		input string
		fn    func(Token, *testing.T)
	}{
		"null": {
			input: "null",
			fn: func(got Token, t *testing.T) {
				assert.Nil(t, got)
			},
		},
		"bool": {
			input: `true`,
			fn: func(got Token, t *testing.T) {
				assert.Equal(t, true, literalValue(t, got))
			},
		},
		"number": {
			input: `42`,
			fn: func(got Token, t *testing.T) {
				assert.EqualValues(t, 42, literalValue(t, got))
			},
		},
		"string": {
			input: `"plain"`,
			fn: func(got Token, t *testing.T) {
				assert.Equal(t, "plain", literalValue(t, got))
			},
		},
		"expression": {
			input: `"${{ github.ref }}"`,
			fn: func(got Token, t *testing.T) {
				expr, ok := Expression(got)
				assert.True(t, ok)
				assert.Equal(t, "${{ github.ref }}", expr)
			},
		},
		"array": {
			input: `["ubuntu-latest","x64"]`,
			fn: func(got Token, t *testing.T) {
				seq, ok := got.(sequenceToken)
				require.True(t, ok)
				assert.Len(t, seq, 2)
				assert.Equal(t, "ubuntu-latest", literalValue(t, seq[0]))
				assert.Equal(t, "x64", literalValue(t, seq[1]))
			},
		},
		"object": {
			input: `{"os":"ubuntu-latest","ref":"${{ github.ref }}"}`,
			fn: func(got Token, t *testing.T) {
				mapping, ok := got.(mappingToken)
				require.True(t, ok)
				assert.Len(t, mapping, 2)
				assert.Equal(t, "os", literalValue(t, mapping[0][0]))
				assert.Equal(t, "ubuntu-latest", literalValue(t, mapping[0][1]))
				assert.Equal(t, "ref", literalValue(t, mapping[1][0]))
				expr, ok := Expression(mapping[1][1])
				assert.True(t, ok)
				assert.Equal(t, "${{ github.ref }}", expr)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn, opt))
	}
}

func TestMappingTokenSerde(t *testing.T) {
	opt := json.WithUnmarshalers(json.UnmarshalFromFunc(unmarshalToken))

	t.Run("object", unmarshal(`{"key":"value"}`, func(got mappingToken, t *testing.T) {
		assert.Len(t, got, 1)
		assert.Equal(t, "key", literalValue(t, got[0][0]))
		assert.Equal(t, "value", literalValue(t, got[0][1]))
	}, opt))

	t.Run("invalid kind", func(t *testing.T) {
		var got mappingToken
		err := json.Unmarshal([]byte(`[]`), &got, opt)

		assert.ErrorContains(t, err, "unknown token type")
	})
}

func literalValue(t *testing.T, token Token) any {
	t.Helper()
	literal, ok := token.(*literalToken)
	if !assert.Truef(t, ok, "got token %T, want *literalToken", token) {
		return nil
	}
	return literal.value
}
