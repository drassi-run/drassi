/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package evaluator

import (
	"slices"
	"strings"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/ast/operators"
)

// When a status function is not referenced, refined "CONDITION" as "success() && <CONDITION>"
type refiner struct {
}

func (c *refiner) Refine(node ast.Node) ast.Node {
	if c.Visit(node) == true {
		return node
	}
	return &ast.OperatorNode{
		Operator: operators.LogicalAnd,
		Operands: []ast.Node{
			&ast.FunctionNode{Name: "success"},
			node,
		},
	}
}

func (c *refiner) Visit(node ast.Node) any {
	return node.Accept(c)
}

func (c *refiner) VisitLiteralNode(node *ast.LiteralNode) any {
	return false
}

func (c *refiner) VisitVariableNode(node *ast.VariableNode) any {
	return false
}

func (c *refiner) VisitOperatorNode(node *ast.OperatorNode) any {
	for _, o := range node.Operands {
		if c.Visit(o) == true {
			return true
		}
	}
	return false
}

func (c *refiner) VisitPropertyAccessNode(node *ast.PropertyAccessNode) any {
	return c.Visit(node.Object)
}

func (c *refiner) VisitIndexAccessNode(node *ast.IndexAccessNode) any {
	if c.Visit(node.Object) == true {
		return true
	}
	for _, i := range node.Indexes {
		if c.Visit(i) == true {
			return true
		}
	}
	return false
}

func (c *refiner) VisitFunctionNode(node *ast.FunctionNode) any {
	fnName := strings.ToLower(node.Name)
	if slices.Contains(statusFns, fnName) {
		return true
	}
	for _, a := range node.Arguments {
		if c.Visit(a) == true {
			return true
		}
	}
	return false
}

var statusFns = []string{"always", "cancelled", "success", "failure"}
