package libraries

import (
	"math"
	"strings"

	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Contains(search ref.Val, item ref.Val) bool {
	// search is an array, check it contains item element
	if search.Type() == ref.TypeList {
		if list, ok := search.(traits.Iterable); ok {
			it := list.Iterator()
			for it.HasNext() {
				_, e := it.Next()
				if EqualWeak(e, item) {
					return true
				}
			}
			return false
		}
	}

	s, ok := search.(traits.Stringable)
	if !ok {
		return false
	}
	i, ok := item.(traits.Stringable)
	if !ok {
		return false
	}

	// Both search, item are convertable to string, check if item is a substring of search.
	// GitHub ignores case when comparing strings.
	ss := strings.ToLower(s.ToString())
	si := strings.ToLower(i.ToString())
	return strings.Contains(ss, si)
}

func EqualWeak(x, y ref.Val) bool {
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
