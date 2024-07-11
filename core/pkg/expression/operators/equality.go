package operators

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

func Equals(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return types.EqualWeak(l, r)
}

func NotEquals(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return !types.EqualWeak(l, r)
}
