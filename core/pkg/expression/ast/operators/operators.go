/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package operators

// Symbolic operators.
const (
	LogicalNot     = "!_"
	LogicalAnd     = "_&&_"
	LogicalOr      = "_||_"
	Equals         = "_==_"
	NotEquals      = "_!=_"
	Less           = "_<_"
	LessEquals     = "_<=_"
	Greater        = "_>_"
	GreaterEquals  = "_>=_"
	PropertyAccess = "_._"
	IndexAccess    = "_[_]"
)

type operatorProp struct {
	symbol     string
	arity      int
	chainable  bool
	precedence int
}

var operatorMap = map[string]operatorProp{
	LogicalNot:     {symbol: "!", arity: 1, chainable: false, precedence: 2},
	LogicalAnd:     {symbol: "&&", arity: 2, chainable: true, precedence: 5},
	LogicalOr:      {symbol: "||", arity: 2, chainable: true, precedence: 6},
	Equals:         {symbol: "==", arity: 2, chainable: false, precedence: 4},
	NotEquals:      {symbol: "!=", arity: 2, chainable: false, precedence: 4},
	Less:           {symbol: "<", arity: 2, chainable: true, precedence: 3},
	LessEquals:     {symbol: "<=", arity: 2, chainable: true, precedence: 3},
	Greater:        {symbol: ">", arity: 2, chainable: true, precedence: 3},
	GreaterEquals:  {symbol: ">=", arity: 2, chainable: true, precedence: 3},
	PropertyAccess: {symbol: ".", arity: 2, chainable: true, precedence: 1},
	IndexAccess:    {symbol: "[]", arity: 2, chainable: true, precedence: 1},
}

// Symbol returns the operator symbol
func Symbol(operator string) string {
	op, found := operatorMap[operator]
	if !found {
		return ""
	}
	return op.symbol
}

// Arity returns the number of argument the operator takes
// -1 is returned if an undefined symbol is provided
func Arity(operator string) int {
	op, found := operatorMap[operator]
	if !found {
		return -1
	}
	return op.arity
}

// Chainable determines the operator is chainable or not, e.g `a && b && c`
func Chainable(operator string) bool {
	op, found := operatorMap[operator]
	return found && op.chainable
}

// Precedence returns the operator precedence, where the higher the number indicates
// higher precedence operations.
func Precedence(operator string) int {
	op, found := operatorMap[operator]
	if !found {
		return 0
	}
	return op.precedence
}
