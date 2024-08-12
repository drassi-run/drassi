package types

import (
	"fmt"
	"math"
	"strconv"

	"drassi.run/core/pkg/expression/types/ref"
)

type Float float64

//goland:noinspection GoSnakeCaseUsage
var (
	NAN          = Float(math.NaN())
	POSITIVE_INF = Float(math.Inf(1))
	NEGATIVE_INF = Float(math.Inf(-1))
)

func IsNaN(v ref.Val) bool {
	f, ok := v.(Float)
	return ok && math.IsNaN(float64(f))
}

func (f Float) Type() ref.Type {
	return ref.TypeFloat
}

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
	// strconv.FormatFloat returns "+Inf" and "-Inf"
	if math.IsInf(float64(f), 1) {
		return "Infinity"
	} else if math.IsInf(float64(f), -1) {
		return "-Infinity"
	}

	// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L34
	return strconv.FormatFloat(float64(f), 'G', 15, 64)
}

func (f Float) Compare(other ref.Val) (int, error) {
	o, ok := other.(Float)
	if !ok {
		return 0, fmt.Errorf("%s vs. %s: %w", f.Type(), other.Type(), errUncomparable)
	}
	if math.IsNaN(float64(f)) || math.IsNaN(float64(o)) {
		return 0, errNaNCompare
	}

	if f < o {
		return -1, nil
	} else if f > o {
		return 1, nil
	} else {
		return 0, nil
	}
}
