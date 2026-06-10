/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package evaluator

import (
	"testing"

	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	. "drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
)

var (
	la  = []any{"abc", true, 3.14}
	ls  = []string{"one", "two", "three"}
	ms  = map[string]int{"first": 1, "second": 2, "third": 3}
	mi  = map[int]any{1: "value", 2: 3.14, 3: false}
	env expression.Env
	ur  *unraveler
)

type staticStatusProvider records.Result

func (s staticStatusProvider) Status() records.Result {
	return records.Result(s)
}

func init() {
	sp := staticStatusProvider(records.ResultSuccess)

	var err error
	env, err = expression.NewEnv(
		expression.WithLibrary(libraries.StdLib()),
		expression.WithLibrary(libraries.StatusLib(sp)),
		expression.WithVariable("la", la),
		expression.WithVariable("ls", ls),
		expression.WithVariable("ms", ms),
		expression.WithVariable("mi", mi),
	)
	if err != nil {
		panic(err)
	}

	ur = &unraveler{env: env}
}

func TestConditional(t *testing.T) {
	t.Run("success", testConditionalSuccess)
	t.Run("failure", testConditionalFailure)
}

func testConditionalSuccess(t *testing.T) {
	sp := staticStatusProvider(records.ResultFailure)

	fEnv, err := env.New(
		expression.WithLibrary(libraries.StatusLib(sp)),
	)
	assert.NoError(t, err)

	tc := []Conditional{
		"",
		"'foobar'",
		"1234",
		"true",
		"la[2]",
		"ms.first",
		"ms['first']",
		"ls",
		"ms",
	}

	for _, c := range tc {
		b1, err := Meet(env, c)
		assert.NoErrorf(t, err, "Conditional: %s", c)
		assert.Truef(t, b1, "Conditional: %s", c)

		b2, err := Meet(fEnv, c)
		assert.NoErrorf(t, err, "Conditional: %s", c)
		assert.Falsef(t, b2, "Conditional: %s", c)
	}
}

func testConditionalFailure(t *testing.T) {
	tc := []Conditional{
		"null",
		"false",
		"0",
		"NaN",
		"''",
	}
	for _, c := range tc {
		b, err := Meet(env, c)
		assert.NoErrorf(t, err, "Conditional: %s", c)
		assert.Falsef(t, b, "Conditional: %s", c)
	}
}
