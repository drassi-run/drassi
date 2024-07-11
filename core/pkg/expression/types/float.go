package types

import (
	"math"
	"strconv"

	"drassi.run/core/pkg/expression/types/ref"
)

type Float float64

func (f Float) Value() any {
	return float64(f)
}

func (f Float) Equal(other ref.Val) bool {
	o, ok := other.(Float)
	return ok && f == o
}

func (f Float) ToBoolean() bool {
	return f != 0 && !math.IsNaN(float64(f))
}

func (f Float) ToNumber() float64 {
	return float64(f)
}

func (f Float) ToString() string {
	return strconv.FormatFloat(float64(f), 'g', -1, 64)
}
