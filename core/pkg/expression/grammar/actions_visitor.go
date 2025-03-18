/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// Code generated from ActionsParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // ActionsParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by ActionsParser.
type ActionsVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by ActionsParser#template.
	VisitTemplate(ctx *TemplateContext) any

	// Visit a parse tree produced by ActionsParser#text.
	VisitText(ctx *TextContext) any

	// Visit a parse tree produced by ActionsParser#placeholder.
	VisitPlaceholder(ctx *PlaceholderContext) any

	// Visit a parse tree produced by ActionsParser#expression.
	VisitExpression(ctx *ExpressionContext) any

	// Visit a parse tree produced by ActionsParser#expr.
	VisitExpr(ctx *ExprContext) any

	// Visit a parse tree produced by ActionsParser#indexAccess.
	VisitIndexAccess(ctx *IndexAccessContext) any

	// Visit a parse tree produced by ActionsParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) any

	// Visit a parse tree produced by ActionsParser#variable.
	VisitVariable(ctx *VariableContext) any

	// Visit a parse tree produced by ActionsParser#propertyAccess.
	VisitPropertyAccess(ctx *PropertyAccessContext) any

	// Visit a parse tree produced by ActionsParser#wrap.
	VisitWrap(ctx *WrapContext) any

	// Visit a parse tree produced by ActionsParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) any

	// Visit a parse tree produced by ActionsParser#literal.
	VisitLiteral(ctx *LiteralContext) any
}
