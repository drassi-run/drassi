/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package types

import (
	"fmt"
	"strconv"

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
	return strconv.FormatBool(bool(b))
}

// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/EvaluationResult.cs#L309
// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/EvaluationResult.cs#L351
func (b Boolean) Compare(other ref.Val) (int, error) {
	o, ok := other.(Boolean)
	if !ok {
		return 0, fmt.Errorf("%w: %s vs. %s", errUncomparable, b.Type(), other.Type())
	}

	if !b && o {
		return -1, nil
	} else if b && !o {
		return 1, nil
	} else {
		return 0, nil
	}
}
