/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"encoding/json/v2"
	"testing"

	"drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunsSerde(t *testing.T) {
	opt := json.WithUnmarshalers(JsonUnmarshalers())

	t.Run("null", func(t *testing.T) {
		var got Action
		input := `{"runs":null}`
		err := json.Unmarshal([]byte(input), &got, opt)
		assert.NoError(t, err)

		assert.Nil(t, got.Runs)
	})

	t.Run("node", func(t *testing.T) {
		var got Action
		input := `{"runs":{"using":"node20","main":"dist/index.js"}}`
		err := json.Unmarshal([]byte(input), &got, opt)
		assert.NoError(t, err)

		runs, ok := got.Runs.(*NodeRuns)
		require.True(t, ok)
		assert.Equal(t, "node20", runs.Using)
		assert.Equal(t, "dist/index.js", runs.Main)
	})

	t.Run("docker", func(t *testing.T) {
		var got Action
		input := `{"runs":{"using":"docker","image":"Dockerfile","args":["--flag"]}}`
		err := json.Unmarshal([]byte(input), &got, opt)
		assert.NoError(t, err)

		runs, ok := got.Runs.(*DockerRuns)
		require.True(t, ok)
		assert.Equal(t, "docker", runs.Using)
		assert.Equal(t, "Dockerfile", runs.Image)
	})

	t.Run("composite", func(t *testing.T) {
		var got Action
		input := `{"runs":{"using":"composite","steps":[{"run":"echo hi"},{"uses":"actions/checkout@v4"}]}}`
		err := json.Unmarshal([]byte(input), &got, opt)
		assert.NoError(t, err)

		runs, ok := got.Runs.(*CompositeRuns)
		require.True(t, ok)
		assert.Len(t, runs.Steps, 2)
		_, isRun := runs.Steps[0].(*workflows.RunActionStep)
		_, isUses := runs.Steps[1].(*workflows.UsesActionStep)
		assert.True(t, isRun)
		assert.True(t, isUses)
	})

	t.Run("unknown", func(t *testing.T) {
		var got Action
		input := `{"runs":{"using":"python3","main":"main.py"}}`
		err := json.Unmarshal([]byte(input), &got, opt)

		assert.ErrorContains(t, err, `unknown runs with using="python3"`)
	})
}
