/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
	"drassi.run/core/pkg/expression/types/traits"
)

func Less(left, right ref.Val) ref.Val {
	r := lessThan(left, right)
	return types.Boolean(r)
}

func LessEquals(left, right ref.Val) ref.Val {
	r := equalWeak(left, right) || lessThan(left, right)
	return types.Boolean(r)
}

func Greater(left, right ref.Val) ref.Val {
	r := greaterThan(left, right)
	return types.Boolean(r)
}

func GreaterEquals(left, right ref.Val) ref.Val {
	r := equalWeak(left, right) || greaterThan(left, right)
	return types.Boolean(r)
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
