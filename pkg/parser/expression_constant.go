package parser

import (
	"math"
)

type ExpressionConstant struct {
	WellKnownFns map[string]IFnInfo[iFn]
}

var e *ExpressionConstant

func NewExpressionConstants() *ExpressionConstant {
	if e != nil {
		return e
	}
	e = &ExpressionConstant{WellKnownFns: make(map[string]IFnInfo[iFn])}
	// WellKnownFns:
	// [x] contains
	// [x] endsWith
	// [x] startsWith
	// [x] join
	// [x] format
	// [] toJson
	// [] fromJson
	e.WellKnownFns["contains"] = NewFunctionInfo[ContainsFn]("contains", 2, 2)
	e.WellKnownFns["endsWith"] = NewFunctionInfo[EndsWithFn]("endsWith", 2, 2)
	e.WellKnownFns["startsWith"] = NewFunctionInfo[StartsWithFn]("startsWith", 2, 2)
	e.WellKnownFns["join"] = NewFunctionInfo[JoinFn]("join", 1, 2)
	e.WellKnownFns["format"] = NewFunctionInfo[FormatFn]("format", 1, math.MaxUint8)
	e.WellKnownFns["fromJson"] = NewFunctionInfo[FromJsonFn]("fromJson", 1, 1)
	return e
}
