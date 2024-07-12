// Code generated from GHAParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // GHAParser
import "github.com/antlr4-go/antlr/v4"

// GHAParserListener is a complete listener for a parse tree produced by GHAParser.
type GHAParserListener interface {
	antlr.ParseTreeListener

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterExpr is called when entering the expr production.
	EnterExpr(c *ExprContext)

	// EnterIndexAccess is called when entering the indexAccess production.
	EnterIndexAccess(c *IndexAccessContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterVariable is called when entering the variable production.
	EnterVariable(c *VariableContext)

	// EnterPropertyAccess is called when entering the propertyAccess production.
	EnterPropertyAccess(c *PropertyAccessContext)

	// EnterWrap is called when entering the wrap production.
	EnterWrap(c *WrapContext)

	// EnterIdentifier is called when entering the identifier production.
	EnterIdentifier(c *IdentifierContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitExpr is called when exiting the expr production.
	ExitExpr(c *ExprContext)

	// ExitIndexAccess is called when exiting the indexAccess production.
	ExitIndexAccess(c *IndexAccessContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitVariable is called when exiting the variable production.
	ExitVariable(c *VariableContext)

	// ExitPropertyAccess is called when exiting the propertyAccess production.
	ExitPropertyAccess(c *PropertyAccessContext)

	// ExitWrap is called when exiting the wrap production.
	ExitWrap(c *WrapContext)

	// ExitIdentifier is called when exiting the identifier production.
	ExitIdentifier(c *IdentifierContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)
}
