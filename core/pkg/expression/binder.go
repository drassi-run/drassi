/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package expression

import (
	"fmt"
	"strings"

	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/ast/operators"
	"drassi.run/core/pkg/expression/types"
	"drassi.run/core/pkg/expression/types/ref"
)

type binder struct {
	env    *env
	stacks []ast.Node
}

// Bind associate ast.Node to the variables, functions and operators implementation in the env
// but defer the execution later.
func (b *binder) Bind(node ast.Node) (ref.LazyVal, error) {
	r := b.Visit(node)
	if err, ok := r.(error); ok {
		return nil, err
	}
	return r.(ref.LazyVal), nil
}

func (b *binder) pushNode(n ast.Node) {
	b.stacks = append(b.stacks, n)
}

func (b *binder) popNode() {
	if len(b.stacks) > 0 {
		b.stacks = b.stacks[:len(b.stacks)-1]
	}
}

func (b *binder) Visit(node ast.Node) any {
	r := node.Accept(b)
	if err, ok := r.(error); ok {
		return err
	}
	val := r.(ref.LazyVal)

	return func() ref.Val {
		b.pushNode(node)
		defer b.popNode()
		return val()
	}
}

func (b *binder) VisitLiteralNode(node *ast.LiteralNode) any {
	return func() ref.Val {
		return node.Value
	}
}

func (b *binder) VisitVariableNode(node *ast.VariableNode) any {
	name := node.Name
	if alias, ok := b.env.alias[name]; ok {
		name = alias
	}

	variable, ok := b.env.variables[name]
	if !ok {
		return fmt.Errorf("undefined variable: %s", node.Name)
	}

	return func() ref.Val {
		return variable
	}
}

func (b *binder) VisitOperatorNode(node *ast.OperatorNode) any {
	operator, ok := b.env.functions[node.Operator]
	if !ok {
		return fmt.Errorf("undefined operator: %q", operators.Symbol(node.Operator))
	}
	minArgs, maxArgs := operator.NumArgs()
	numArgs := len(node.Operands)
	if numArgs < minArgs {
		return fmt.Errorf("not enough args for operator %q, min %d, got %d", operators.Symbol(node.Operator), minArgs, numArgs)
	}
	if numArgs > maxArgs {
		return fmt.Errorf("too many args for operator %q, max %d, got %d", operators.Symbol(node.Operator), maxArgs, numArgs)
	}

	operands := make([]ref.LazyVal, len(node.Operands))
	for i, operand := range node.Operands {
		r := b.Visit(operand)
		if err, ok := r.(error); ok {
			return err
		}
		operands[i] = r.(ref.LazyVal)
	}

	return operator.Bind(operands...)
}

func (b *binder) VisitPropertyAccessNode(node *ast.PropertyAccessNode) any {
	operator, ok := b.env.functions[operators.PropertyAccess]
	if !ok {
		return fmt.Errorf("undefined operator: %q", operators.Symbol(operators.PropertyAccess))
	}

	r := b.Visit(node.Object)
	if err, ok := r.(error); ok {
		return err
	}
	object := r.(ref.LazyVal)

	properties := make([]ref.LazyVal, len(node.Properties))
	for i, prop := range node.Properties {
		p := types.String(prop)
		properties[i] = func() ref.Val {
			return p
		}
	}

	return operator.Bind(prepend(properties, object)...)
}

func (b *binder) VisitIndexAccessNode(node *ast.IndexAccessNode) any {
	operator, ok := b.env.functions[operators.IndexAccess]
	if !ok {
		return fmt.Errorf("undefined operator %q", operators.Symbol(operators.IndexAccess))
	}

	r := b.Visit(node.Object)
	if err, ok := r.(error); ok {
		return err
	}
	object := r.(ref.LazyVal)

	indexes := make([]ref.LazyVal, len(node.Indexes))
	for i, idx := range node.Indexes {
		r := b.Visit(idx)
		if err, ok := r.(error); ok {
			return err
		}
		indexes[i] = r.(ref.LazyVal)
	}

	return operator.Bind(prepend(indexes, object)...)
}

func (b *binder) VisitFunctionNode(node *ast.FunctionNode) any {
	fnName := strings.ToLower(node.Name)
	function, ok := b.env.functions[fnName]
	if !ok {
		return fmt.Errorf("undefined function: %s", node.Name)
	}
	minArgs, maxArgs := function.NumArgs()
	numArgs := len(node.Arguments)
	if numArgs < minArgs {
		return fmt.Errorf("not enough args for function %q, min %d, got %d", node.Name, minArgs, numArgs)
	}
	if numArgs > maxArgs {
		return fmt.Errorf("too many args for function %q, max %d, got %d", node.Name, maxArgs, numArgs)
	}

	arguments := make([]ref.LazyVal, numArgs)
	for i, arg := range node.Arguments {
		r := b.Visit(arg)
		if err, ok := r.(error); ok {
			return err
		}
		arguments[i] = r.(ref.LazyVal)
	}

	return function.Bind(arguments...)
}

func prepend[T any](slice []T, elems ...T) []T {
	return append(elems, slice...)
}
