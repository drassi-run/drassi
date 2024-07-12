// Code generated from GHAParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // GHAParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by GHAParser.
type GHAParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by GHAParser#expression.
	VisitExpression(ctx *ExpressionContext) any

	// Visit a parse tree produced by GHAParser#expr.
	VisitExpr(ctx *ExprContext) any

	// Visit a parse tree produced by GHAParser#indexAccess.
	VisitIndexAccess(ctx *IndexAccessContext) any

	// Visit a parse tree produced by GHAParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) any

	// Visit a parse tree produced by GHAParser#variable.
	VisitVariable(ctx *VariableContext) any

	// Visit a parse tree produced by GHAParser#propertyAccess.
	VisitPropertyAccess(ctx *PropertyAccessContext) any

	// Visit a parse tree produced by GHAParser#wrap.
	VisitWrap(ctx *WrapContext) any

	// Visit a parse tree produced by GHAParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) any

	// Visit a parse tree produced by GHAParser#literal.
	VisitLiteral(ctx *LiteralContext) any
}
