package types

import (
	"fmt"

	"drassi.run/core/pkg/expression/types/ref"
)

type Boolean bool

const (
	TRUE  = Boolean(true)
	FALSE = Boolean(false)
)

func (b Boolean) Type() ref.Type {
	return ref.TypeBoolean
}

func (b Boolean) Value() any {
	return bool(b)
}

func (b Boolean) Equal(other ref.Val) bool {
	o, ok := other.(Boolean)
	return ok && b == o
}

func (b Boolean) ToBoolean() bool {
	return bool(b)
}

func (b Boolean) ToNumber() float64 {
	if b {
		return 1
	}
	return 0
}

func (b Boolean) ToString() string {
	if b {
		return "true"
	}
	return "false"
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/EvaluationResult.cs#L309
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/EvaluationResult.cs#L351
func (b Boolean) Compare(other ref.Val) (int, error) {
	o, ok := other.(Boolean)
	if !ok {
		return 0, fmt.Errorf("cannot compare non-boolean types")
	}

	if !b && o {
		return -1, nil
	} else if b && !o {
		return 1, nil
	} else {
		return 0, nil
	}
}
