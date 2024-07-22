// Code generated from Actions.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // Actions
import "github.com/antlr4-go/antlr/v4"

type BaseActionsVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseActionsVisitor) VisitExpression(ctx *ExpressionContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitExpr(ctx *ExprContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitIndexAccess(ctx *IndexAccessContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitFunctionCall(ctx *FunctionCallContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitVariable(ctx *VariableContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitPropertyAccess(ctx *PropertyAccessContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitWrap(ctx *WrapContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitIdentifier(ctx *IdentifierContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseActionsVisitor) VisitLiteral(ctx *LiteralContext) any {
	return v.VisitChildren(ctx)
}
