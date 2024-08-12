package libraries

import (
	"math"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func equalWeak(x, y ref.Val) bool {
	if x.Type() == y.Type() {
		return x.Equal(y)
	}

	fx := coerce(x)
	fy := coerce(y)
	return fx == fy
}

func coerce(x ref.Val) float64 {
	if f, ok := x.(traits.Numerical); ok {
		return f.ToNumber()
	}
	return math.NaN()
}
