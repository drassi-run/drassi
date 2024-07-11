package types

import (
	"drassi.run/core/pkg/expression/types/ref"
)

type Null struct{}

var NULL = Null{}

func (n Null) Value() any {
	return nil
}

func (n Null) Equal(other ref.Val) bool {
	_, ok := other.(Null)
	return ok
}

func (n Null) ToBoolean() bool {
	return false
}

func (n Null) ToNumber() float64 {
	return 0
}

func (n Null) ToString() string {
	return ""
}
