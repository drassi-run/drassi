package evaluator

import (
	"drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/libraries"
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
	var err error
	env, err = expression.NewEnv(
		expression.WithLibrary(libraries.StdLib()),
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
	t.Run("success", func(t *testing.T) {
		tc := []Conditional{
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
			b, err := Meet(env, c)
			assert.NoErrorf(t, err, "Conditional: %s", c)
			assert.Truef(t, b, "Conditional: %s", c)
		}
	})

	t.Run("failure", func(t *testing.T) {
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
	})
}
