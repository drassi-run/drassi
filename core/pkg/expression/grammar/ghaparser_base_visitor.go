// Code generated from GHAParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // GHAParser
import "github.com/antlr4-go/antlr/v4"

type BaseGHAParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseGHAParserVisitor) VisitExpression(ctx *ExpressionContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitExpr(ctx *ExprContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitIndexAccess(ctx *IndexAccessContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitFunctionCall(ctx *FunctionCallContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitVariable(ctx *VariableContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitPropertyAccess(ctx *PropertyAccessContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitWrap(ctx *WrapContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitIdentifier(ctx *IdentifierContext) any {
	return v.VisitChildren(ctx)
}

func (v *BaseGHAParserVisitor) VisitLiteral(ctx *LiteralContext) any {
	return v.VisitChildren(ctx)
}
