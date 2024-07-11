package types

import (
	"math"
	"strconv"
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
)

type String string

func (s String) Value() any {
	return string(s)
}

func (s String) Equal(other ref.Val) bool {
	o, ok := other.(String)
	// GitHub ignores case when comparing strings.
	return ok && strings.EqualFold(string(s), string(o))
}

func (s String) ToBoolean() bool {
	return len(s) > 0
}

func (s String) ToNumber() float64 {
	// empty string returns 0
	if len(s) == 0 {
		return 0
	}

	// parsed from any legal JSON number format
	if f, err := strconv.ParseFloat(string(s), 64); err == nil {
		return f
	}

	// otherwise NaN
	return math.NaN()
}

func (s String) ToString() string {
	return string(s)
}
