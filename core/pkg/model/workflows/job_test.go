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

func TestJobSerde(t *testing.T) {
	opt := json.WithUnmarshalers(JsonUnmarshalers())

	testcases := map[string]struct {
		input string
		fn    func(Job, *testing.T)
	}{
		"null": {
			input: `null`,
			fn: func(got Job, t *testing.T) {
				assert.Nil(t, got)
			},
		},
		"normal job": {
			input: `{"runs-on":"ubuntu-latest","steps":[{"run":"go test ./..."}]}`,
			fn: func(got Job, t *testing.T) {
				job, ok := got.(*NormalJob)
				require.True(t, ok)
				assert.Equal(t, "ubuntu-latest", literalValue(t, job.RunsOn))
				assert.Len(t, job.Steps, 1)
			},
		},
		"normal job runs-on array": {
			input: `{"runs-on":["self-hosted","linux","x64"],"steps":[{"run":"go test ./..."}]}`,
			fn: func(got Job, t *testing.T) {
				job, ok := got.(*NormalJob)
				require.True(t, ok)
				seq, ok := job.RunsOn.(sequenceToken)
				require.True(t, ok)
				assert.Len(t, seq, 3)
				assert.Equal(t, "self-hosted", literalValue(t, seq[0]))
				assert.Len(t, job.Steps, 1)
			},
		},
		"normal job runs-on object": {
			input: `{"runs-on":{"group":"gpu-runner","labels":["self-hosted","x64"]},"steps":[{"run":"go test ./..."}]}`,
			fn: func(got Job, t *testing.T) {
				job, ok := got.(*NormalJob)
				require.True(t, ok)
				mapping, ok := job.RunsOn.(mappingToken)
				require.True(t, ok)
				assert.Len(t, mapping, 2)
				assert.Equal(t, "group", literalValue(t, mapping[0][0]))
				assert.Equal(t, "gpu-runner", literalValue(t, mapping[0][1]))
				assert.Len(t, job.Steps, 1)
			},
		},
		"reusable workflow call job": {
			input: `{"uses":"./.github/workflows/deploy.yml"}`,
			fn: func(got Job, t *testing.T) {
				job, ok := got.(*ReusableWorkflowCallJob)
				require.True(t, ok)
				assert.Equal(t, "./.github/workflows/deploy.yml", job.Uses)
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn, opt))
	}

	t.Run("both runs-on and uses", func(t *testing.T) {
		var got Job
		err := json.Unmarshal([]byte(`{"runs-on":"ubuntu-latest","uses":"./workflow.yml"}`), &got, opt)

		assert.ErrorContains(t, err, "either `runs-on` or `uses`")
	})

	t.Run("missing runs-on and uses", func(t *testing.T) {
		var got Job
		err := json.Unmarshal([]byte(`{"name":"bad"}`), &got, opt)

		assert.ErrorContains(t, err, "either `runs-on` or `uses`")
	})
}

func TestArraySerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(array, *testing.T)
	}{
		"string": {
			input: `"build"`,
			fn: func(got array, t *testing.T) {
				assert.EqualValues(t, array{"build"}, got)
			},
		},
		"array": {
			input: `["build","test"]`,
			fn: func(got array, t *testing.T) {
				assert.EqualValues(t, array{"build", "test"}, got)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got array
		err := json.Unmarshal([]byte(`true`), &got)

		assert.ErrorContains(t, err, "expected string or array")
	})
}

func TestJobSecretsSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(JobSecrets, *testing.T)
	}{
		"inherit": {
			input: `"inherit"`,
			fn: func(got JobSecrets, t *testing.T) {
				assert.True(t, got.Inherit)
				assert.Nil(t, got.Secrets)
			},
		},
		"object": {
			input: `{"MESSAGE":"Hello world"}`,
			fn: func(got JobSecrets, t *testing.T) {
				assert.False(t, got.Inherit)
				assert.EqualValues(t, map[string]string{"MESSAGE": "Hello world"}, got.Secrets)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("unexpected string", func(t *testing.T) {
		var got JobSecrets
		err := json.Unmarshal([]byte(`"all"`), &got)

		assert.ErrorContains(t, err, "unexpected JobSecrets=all")
	})

	t.Run("invalid kind", func(t *testing.T) {
		var got JobSecrets
		err := json.Unmarshal([]byte(`[]`), &got)

		assert.ErrorContains(t, err, "expected object for JobSecrets")
	})
}

func TestEnvironmentSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(Environment, *testing.T)
	}{
		"string shorthand": {
			input: `"production"`,
			fn: func(got Environment, t *testing.T) {
				assert.Equal(t, "production", got.Name)
			},
		},
		"object": {
			input: `{"name":"production","url":"https://example.com"}`,
			fn: func(got Environment, t *testing.T) {
				assert.Equal(t, "production", got.Name)
				assert.Equal(t, "https://example.com", got.Url)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got Environment
		err := json.Unmarshal([]byte(`42`), &got)

		assert.ErrorContains(t, err, "expected string or object for Environment")
	})
}

func TestRunsOnSerde(t *testing.T) {
	testcases := map[string]struct {
		input string
		fn    func(RunsOn, *testing.T)
	}{
		"string shorthand": {
			input: `"ubuntu-latest"`,
			fn: func(got RunsOn, t *testing.T) {
				assert.EqualValues(t, array{"ubuntu-latest"}, got.Labels)
			},
		},
		"array shorthand": {
			input: `["self-hosted","x64"]`,
			fn: func(got RunsOn, t *testing.T) {
				assert.EqualValues(t, array{"self-hosted", "x64"}, got.Labels)
			},
		},
		"object": {
			input: `{"group":"linux","labels":["self-hosted","x64"]}`,
			fn: func(got RunsOn, t *testing.T) {
				assert.Equal(t, "linux", got.Group)
				assert.EqualValues(t, array{"self-hosted", "x64"}, got.Labels)
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn))
	}

	t.Run("invalid kind", func(t *testing.T) {
		var got RunsOn
		err := json.Unmarshal([]byte(`true`), &got)

		assert.ErrorContains(t, err, "expected string, array or object for RunsOn")
	})
}
