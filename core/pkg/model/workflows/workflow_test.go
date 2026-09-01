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
)

func TestOnSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(On, *testing.T)
	}{
		"string shorthand": {
			input: `"push"`,
			fn: func(got On, t *testing.T) {
				_, ok := got["push"]
				assert.True(t, ok)
				assert.Len(t, got, 1)
			},
		},
		"array shorthand": {
			input: `["push","pull_request"]`,
			fn: func(got On, t *testing.T) {
				_, hasPush := got["push"]
				_, hasPullRequest := got["pull_request"]
				assert.True(t, hasPush)
				assert.True(t, hasPullRequest)
				assert.Len(t, got, 2)
			},
		},
		"object": {
			input: `{"pull_request":{"types":["opened"]}}`,
			fn: func(got On, t *testing.T) {
				assert.Len(t, got, 1)
				assert.JSONEq(t, `{"types":["opened"]}`, string(got["pull_request"]))
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got On
		input := `true`
		err := json.Unmarshal([]byte(input), &got)
		assert.ErrorContains(t, err, "expected string, array or object for On")
	})
}

func TestPermissionsSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(Permissions, *testing.T)
	}{
		"read-all shorthand": {
			input: `"read-all"`,
			fn: func(got Permissions, t *testing.T) {
				assert.Equal(t, Permissions{"*": PermissionsLevelRead}, got)
			},
		},
		"write-all shorthand": {
			input: `"write-all"`,
			fn: func(got Permissions, t *testing.T) {
				assert.Equal(t, Permissions{"*": PermissionsLevelWrite}, got)
			},
		},
		"object": {
			input: `{"actions":"read","contents":"write","issues":"none"}`,
			fn: func(got Permissions, t *testing.T) {
				assert.Equal(t, Permissions{
					"actions":  PermissionsLevelRead,
					"contents": PermissionsLevelWrite,
					"issues":   PermissionsLevelNone,
				}, got)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("unknown shorthand", func(t *testing.T) {
		var got Permissions
		err := json.Unmarshal([]byte(`"admin-all"`), &got)

		assert.ErrorContains(t, err, "unknown permission admin-all")
	})

	t.Run("invalid kind", func(t *testing.T) {
		var got Permissions
		err := json.Unmarshal([]byte(`true`), &got)

		assert.ErrorContains(t, err, "expected string or object for Permission")
	})
}

func TestConcurrencySerde(t *testing.T) {
	opt := json.WithUnmarshalers(json.UnmarshalFromFunc(unmarshalToken))

	testcases := map[string]struct {
		input string
		fn    func(Concurrency, *testing.T)
	}{
		"string shorthand": {
			input: `"ci-main"`,
			fn: func(got Concurrency, t *testing.T) {
				assert.Equal(t, "ci-main", literalValue(t, got.Group))
				assert.False(t, got.CancelInProgress)
			},
		},
		"object": {
			input: `{"group":"ci-main","cancel-in-progress":true}`,
			fn: func(got Concurrency, t *testing.T) {
				assert.Equal(t, "ci-main", literalValue(t, got.Group))
				assert.True(t, got.CancelInProgress)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn, opt))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got Concurrency
		err := json.Unmarshal([]byte(`42`), &got, opt)

		assert.ErrorContains(t, err, "expected string or object for Concurrency")
	})
}
