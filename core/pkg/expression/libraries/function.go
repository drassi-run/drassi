/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package libraries

import (
	"math"

	expr "drassi.run/core/pkg/expression"
	"drassi.run/core/pkg/expression/types/ref"
)

// nullaryFn is the function take 0 arguments
type nullaryFn func() ref.Val

var _ expr.Function = (nullaryFn)(nil)

func (f nullaryFn) NumArgs() (min int, max int) {
	return 0, 0
}

func (f nullaryFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f
}

// unaryFn is the function take 1 argument
type unaryFn func(ref.Val) ref.Val

var _ expr.Function = (unaryFn)(nil)

func (f unaryFn) NumArgs() (min int, max int) {
	return 1, 1
}

func (f unaryFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f.invoke(args[0])
}

func (f unaryFn) invoke(arg ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		v := arg()
		if ref.IsError(v) {
			return v
		}

		return f(v)
	}
}

// binaryFn is the function take 2 arguments
type binaryFn func(ref.Val, ref.Val) ref.Val

var _ expr.Function = (binaryFn)(nil)

func (f binaryFn) NumArgs() (min int, max int) {
	return 2, 2
}

func (f binaryFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f.invoke(args[0], args[1])
}

func (f binaryFn) invoke(a1, a2 ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		v1 := a1()
		if ref.IsError(v1) {
			return v1
		}

		v2 := a2()
		if ref.IsError(v2) {
			return v2
		}

		return f(v1, v2)
	}
}

// ternaryFn is the function take 3 arguments
type ternaryFn func(ref.Val, ref.Val, ref.Val) ref.Val

var _ expr.Function = (ternaryFn)(nil)

func (f ternaryFn) NumArgs() (min int, max int) {
	return 3, 3
}

func (f ternaryFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f.invoke(args[0], args[1], args[3])
}

func (f ternaryFn) invoke(a1, a2, a3 ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		v1 := a1()
		if ref.IsError(v1) {
			return v1
		}

		v2 := a2()
		if ref.IsError(v2) {
			return v2
		}

		v3 := a3()
		if ref.IsError(v3) {
			return v3
		}

		return f(v1, v2, v3)
	}
}

// variadicLazyFn is the function take 0 or more arguments
type variadicLazyFn func(...ref.LazyVal) ref.Val

var _ expr.Function = (variadicLazyFn)(nil)

func (f variadicLazyFn) NumArgs() (min int, max int) {
	return 0, math.MaxInt32
}

func (f variadicLazyFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f.invoke(args...)
}

func (f variadicLazyFn) invoke(args ...ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		return f(args...)
	}
}

// oneRestLazyFn is the function take 1 or more arguments
type oneRestLazyFn func(ref.LazyVal, ...ref.LazyVal) ref.Val

var _ expr.Function = (oneRestLazyFn)(nil)

func (f oneRestLazyFn) NumArgs() (min int, max int) {
	return 1, math.MaxInt32
}

func (f oneRestLazyFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	return f.invoke(args[0], args[1:]...)
}

func (f oneRestLazyFn) invoke(a ref.LazyVal, rest ...ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		return f(a, rest...)
	}
}

// oneRestLazyFn is the function take 1 argument and 1 optional argument
type oneOptionFn func(ref.Val, ref.LazyVal) ref.Val

var _ expr.Function = (oneOptionFn)(nil)

func (f oneOptionFn) NumArgs() (min int, max int) {
	return 1, 2
}

func (f oneOptionFn) Bind(args ...ref.LazyVal) ref.LazyVal {
	var a1 ref.LazyVal = args[0]
	var a2 ref.LazyVal = nil
	if len(args) > 1 {
		a2 = args[1]
	}
	return f.invoke(a1, a2)
}

func (f oneOptionFn) invoke(a1, a2 ref.LazyVal) ref.LazyVal {
	return func() ref.Val {
		v1 := a1()
		if ref.IsError(v1) {
			return v1
		}

		return f(v1, a2)
	}
}
