package converter

import (
	"fmt"
	"maps"

	"drassi.run/core/pkg/model/workflows"
)

type multiMapToken struct {
	tokens []workflows.Token
}

func (t *multiMapToken) Unravel(name string, supplier workflows.EvaluationSupplier) (any, error) {
	r := make(map[string]any)

	for _, token := range t.tokens {
		res, err := token.Unravel(name, supplier)
		if err != nil {
			return nil, err
		}
		if m, ok := res.(map[string]any); ok {
			maps.Copy(r, m)
		}
		return nil, fmt.Errorf("expected token return a map[string]any, got %T", res)
	}

	return r, nil
}
