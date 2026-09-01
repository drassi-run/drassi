/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package workflows

import (
	"encoding/json/v2"
	"testing"

	"drassi.run/core/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestStepSerde(t *testing.T) {
	opt := json.WithUnmarshalers(
		json.JoinUnmarshalers(
			model.UnmarshalInterface(discriminateStep),
			json.UnmarshalFromFunc(unmarshalToken),
		),
	)

	testcases := map[string]struct {
		input string
		fn    func(Step, *testing.T)
	}{
		"null": {
			input: `null`,
			fn: func(got Step, t *testing.T) {
				assert.Nil(t, got)
			},
		},
		"run step": {
			input: `{"name":"test","run":"go test ./..."}`,
			fn: func(got Step, t *testing.T) {
				step, ok := got.(*RunActionStep)
				assert.True(t, ok)
				assert.Equal(t, "go test ./...", literalValue(t, step.Run))
				assert.Equal(t, "test", literalValue(t, step.Name))
			},
		},
		"uses step": {
			input: `{"name":"checkout","uses":"actions/checkout@v4"}`,
			fn: func(got Step, t *testing.T) {
				step, ok := got.(*UsesActionStep)
				assert.True(t, ok)
				assert.Equal(t, "actions/checkout@v4", step.Uses)
				assert.Equal(t, "checkout", literalValue(t, step.Name))
			},
		},
	}
	for name, tc := range testcases {
		t.Run(name, unmarshal(tc.input, tc.fn, opt))
	}

	t.Run("both run and uses", func(t *testing.T) {
		var got []Step
		err := json.Unmarshal([]byte(`[{"run":"go test ./...","uses":"actions/checkout@v4"}]`), &got, opt)

		assert.ErrorContains(t, err, "either `run` or `uses`")
	})

	t.Run("missing run and uses", func(t *testing.T) {
		var got []Step
		err := json.Unmarshal([]byte(`[{"name":"missing"}]`), &got, opt)

		assert.ErrorContains(t, err, "either `run` or `uses`")
	})
}
