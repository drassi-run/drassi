package types

import (
	"drassi.run/core/pkg/expression/types/ref"
)

type Boolean bool

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
