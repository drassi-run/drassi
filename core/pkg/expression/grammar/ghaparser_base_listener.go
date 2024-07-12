// Code generated from GHAParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // GHAParser
import "github.com/antlr4-go/antlr/v4"

// BaseGHAParserListener is a complete listener for a parse tree produced by GHAParser.
type BaseGHAParserListener struct{}

var _ GHAParserListener = &BaseGHAParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseGHAParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseGHAParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseGHAParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseGHAParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseGHAParserListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseGHAParserListener) ExitExpression(ctx *ExpressionContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseGHAParserListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseGHAParserListener) ExitExpr(ctx *ExprContext) {}

// EnterIndexAccess is called when production indexAccess is entered.
func (s *BaseGHAParserListener) EnterIndexAccess(ctx *IndexAccessContext) {}

// ExitIndexAccess is called when production indexAccess is exited.
func (s *BaseGHAParserListener) ExitIndexAccess(ctx *IndexAccessContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseGHAParserListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseGHAParserListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterVariable is called when production variable is entered.
func (s *BaseGHAParserListener) EnterVariable(ctx *VariableContext) {}

// ExitVariable is called when production variable is exited.
func (s *BaseGHAParserListener) ExitVariable(ctx *VariableContext) {}

// EnterPropertyAccess is called when production propertyAccess is entered.
func (s *BaseGHAParserListener) EnterPropertyAccess(ctx *PropertyAccessContext) {}

// ExitPropertyAccess is called when production propertyAccess is exited.
func (s *BaseGHAParserListener) ExitPropertyAccess(ctx *PropertyAccessContext) {}

// EnterWrap is called when production wrap is entered.
func (s *BaseGHAParserListener) EnterWrap(ctx *WrapContext) {}

// ExitWrap is called when production wrap is exited.
func (s *BaseGHAParserListener) ExitWrap(ctx *WrapContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseGHAParserListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseGHAParserListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseGHAParserListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseGHAParserListener) ExitLiteral(ctx *LiteralContext) {}
