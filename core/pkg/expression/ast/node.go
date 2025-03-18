/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ast

import "drassi.run/core/pkg/expression/types/ref"

type Visitor interface {
	Visit(node Node) any
	VisitLiteralNode(node *LiteralNode) any
	VisitVariableNode(node *VariableNode) any
	VisitOperatorNode(node *OperatorNode) any
	VisitPropertyAccessNode(node *PropertyAccessNode) any
	VisitIndexAccessNode(node *IndexAccessNode) any
	VisitFunctionNode(node *FunctionNode) any
}

type Node interface {
	Accept(visitor Visitor) any
}

var (
	_ Node = (*LiteralNode)(nil)
	_ Node = (*VariableNode)(nil)
	_ Node = (*OperatorNode)(nil)
	_ Node = (*PropertyAccessNode)(nil)
	_ Node = (*IndexAccessNode)(nil)
	_ Node = (*FunctionNode)(nil)
)

type LiteralNode struct {
	Value ref.Val
}

func (n *LiteralNode) Accept(visitor Visitor) any {
	return visitor.VisitLiteralNode(n)
}

type VariableNode struct {
	Name string
}

func (n *VariableNode) Accept(visitor Visitor) any {
	return visitor.VisitVariableNode(n)
}

type OperatorNode struct {
	Operator string
	Operands []Node
}

func (n *OperatorNode) Accept(visitor Visitor) any {
	return visitor.VisitOperatorNode(n)
}

type PropertyAccessNode struct {
	Object     Node
	Properties []string
}

func (n *PropertyAccessNode) Accept(visitor Visitor) any {
	return visitor.VisitPropertyAccessNode(n)
}

type IndexAccessNode struct {
	Object  Node
	Indexes []Node
}

func (n *IndexAccessNode) Accept(visitor Visitor) any {
	return visitor.VisitIndexAccessNode(n)
}

type FunctionNode struct {
	Name      string
	Arguments []Node
}

func (n *FunctionNode) Accept(visitor Visitor) any {
	return visitor.VisitFunctionNode(n)
}
