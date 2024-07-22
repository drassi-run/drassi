// Code generated from Actions.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // Actions
import "github.com/antlr4-go/antlr/v4"

// BaseActionsListener is a complete listener for a parse tree produced by ActionsParser.
type BaseActionsListener struct{}

var _ ActionsListener = &BaseActionsListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseActionsListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseActionsListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseActionsListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseActionsListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseActionsListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseActionsListener) ExitExpression(ctx *ExpressionContext) {}

// EnterExpr is called when production expr is entered.
func (s *BaseActionsListener) EnterExpr(ctx *ExprContext) {}

// ExitExpr is called when production expr is exited.
func (s *BaseActionsListener) ExitExpr(ctx *ExprContext) {}

// EnterIndexAccess is called when production indexAccess is entered.
func (s *BaseActionsListener) EnterIndexAccess(ctx *IndexAccessContext) {}

// ExitIndexAccess is called when production indexAccess is exited.
func (s *BaseActionsListener) ExitIndexAccess(ctx *IndexAccessContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseActionsListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseActionsListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterVariable is called when production variable is entered.
func (s *BaseActionsListener) EnterVariable(ctx *VariableContext) {}

// ExitVariable is called when production variable is exited.
func (s *BaseActionsListener) ExitVariable(ctx *VariableContext) {}

// EnterPropertyAccess is called when production propertyAccess is entered.
func (s *BaseActionsListener) EnterPropertyAccess(ctx *PropertyAccessContext) {}

// ExitPropertyAccess is called when production propertyAccess is exited.
func (s *BaseActionsListener) ExitPropertyAccess(ctx *PropertyAccessContext) {}

// EnterWrap is called when production wrap is entered.
func (s *BaseActionsListener) EnterWrap(ctx *WrapContext) {}

// ExitWrap is called when production wrap is exited.
func (s *BaseActionsListener) ExitWrap(ctx *WrapContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseActionsListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseActionsListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseActionsListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseActionsListener) ExitLiteral(ctx *LiteralContext) {}
