package libraries

import (
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Less(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return lessThan(l, r)
}

func LessEquals(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return EqualWeak(l, r) || lessThan(l, r)
}

func Greater(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return lessThan(l, r)
}

func GreaterEquals(lhs ref.LazyVal, rhs ref.LazyVal) bool {
	l, r := lhs(), rhs()
	return EqualWeak(l, r) || greaterThan(l, r)
}

func lessThan(lhs ref.Val, rhs ref.Val) bool {
	if lhs.Type() != rhs.Type() {
		lf, rf := coerce(lhs), coerce(rhs)
		return lf < rf
	}

	if c, ok := lhs.(traits.Comparable); ok {
		if v, err := c.Compare(rhs); err != nil {
			return false
		} else {
			return v < 0
		}
	}

	return false
}

func greaterThan(lhs ref.Val, rhs ref.Val) bool {
	if lhs.Type() != rhs.Type() {
		lf, rf := coerce(lhs), coerce(rhs)
		return lf > rf
	}

	if c, ok := lhs.(traits.Comparable); ok {
		if v, err := c.Compare(rhs); err != nil {
			return false
		} else {
			return v > 0
		}
	}

	return false
}
