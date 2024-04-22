package parser

import (
	"math"

	"github.com/dungdm93/drasi/pkg/expression/parser/functions"
)

type ExpressionConstant struct {
	WellKnownFns map[string]functions.IFnInfo[functions.IFn]
}

var e *ExpressionConstant

func NewExpressionConstants() *ExpressionConstant {
	if e != nil {
		return e
	}
	e = &ExpressionConstant{WellKnownFns: make(map[string]functions.IFnInfo[functions.IFn])}
	// WellKnownFns:
	// [x] contains
	// [x] endsWith
	// [x] startsWith
	// [x] join
	// [x] format
	// [] toJson
	// [] fromJson
	e.WellKnownFns["contains"] = functions.NewFunctionInfo[functions.ContainsFn]("contains", 2, 2)
	e.WellKnownFns["endsWith"] = functions.NewFunctionInfo[functions.EndsWithFn]("endsWith", 2, 2)
	e.WellKnownFns["startsWith"] = functions.NewFunctionInfo[functions.StartsWithFn]("startsWith", 2, 2)
	e.WellKnownFns["join"] = functions.NewFunctionInfo[functions.JoinFn]("join", 1, 2)
	e.WellKnownFns["format"] = functions.NewFunctionInfo[functions.FormatFn]("format", 1, math.MaxUint8)
	e.WellKnownFns["fromJson"] = functions.NewFunctionInfo[functions.FromJsonFn]("fromJson", 1, 1)
	return e
}
