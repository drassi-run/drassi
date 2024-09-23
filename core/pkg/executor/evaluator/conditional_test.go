package evaluator

import (
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
	"drassi.run/core/pkg/model/records"
	. "drassi.run/core/pkg/model/workflows"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	la  = []any{"abc", true, 3.14}
	ls  = []string{"one", "two", "three"}
	ms  = map[string]int{"first": 1, "second": 2, "third": 3}
	mi  = map[int]any{1: "value", 2: 3.14, 3: false}
	env expression.Env
	ur  *unraveler
)

func init() {
	job := new(records.Job)
	job.Result = records.ResultSuccess

	var err error
	env, err = expression.NewEnv(
		expression.WithLibrary(libraries.StdLib()),
		expression.WithLibrary(libraries.JobLib(job)),
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
	fJob := new(records.Job)
	fJob.Result = records.ResultFailure

	fEnv, err := env.New(
		expression.WithLibrary(libraries.JobLib(fJob)),
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
