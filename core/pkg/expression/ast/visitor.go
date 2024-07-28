package ast

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"strconv"
	"strings"

	"drassi.run/core/pkg/expression/ast/operators"
	"drassi.run/core/pkg/expression/grammar"
	"drassi.run/core/pkg/expression/types"
)

var tokenOp = map[int]string{
	grammar.ActionsLexerNOT:      operators.LogicalNot,
	grammar.ActionsLexerAND:      operators.LogicalAnd,
	grammar.ActionsLexerOR:       operators.LogicalOr,
	grammar.ActionsLexerLT:       operators.Less,
	grammar.ActionsLexerLTEQ:     operators.LessEquals,
	grammar.ActionsLexerGT:       operators.Greater,
	grammar.ActionsLexerGTEQ:     operators.GreaterEquals,
	grammar.ActionsLexerEQUAL:    operators.Equals,
	grammar.ActionsLexerNOTEQUAL: operators.NotEquals,
}

type astVisitor struct {
	grammar.BaseActionsVisitor
}

func (v *astVisitor) Visit(tree antlr.ParseTree) any { return tree.Accept(v) }

func (v *astVisitor) VisitExpression(ctx *grammar.ExpressionContext) any {
	return v.Visit(ctx.GetE())
}

func (v *astVisitor) VisitExpr(ctx *grammar.ExprContext) any {
	if ea := ctx.ExprAccess(); ea != nil {
		return v.Visit(ea)
	}
	if lit := ctx.Literal(); lit != nil {
		return v.Visit(lit)
	}
	if opToken := ctx.GetOp(); opToken != nil {
		op, ok := tokenOp[opToken.GetTokenType()]
		if !ok {
			return fmt.Errorf("unknown operator %q", opToken.GetText())
		}
		exprs := ctx.AllExpr()
		numExpr := len(exprs)
		if arity := operators.Arity(op); numExpr < arity {
			return fmt.Errorf("not enough operands for operator %q, expect %d, got %d", opToken.GetText(), arity, numExpr)
		} else if !operators.Chainable(op) && numExpr > arity {
			return fmt.Errorf("too many operands for operator %q, expect most %d, got %d", opToken.GetText(), arity, numExpr)
		}
		operands := make([]Node, numExpr)
		for i, expr := range exprs {
			r := v.Visit(expr)
			if err, ok := r.(error); ok {
				return err
			}
			operands[i] = r.(Node)
		}
		return &OperatorNode{
			Operator: op,
			Operands: operands,
		}
	}

	return fmt.Errorf("unknown expression %q", ctx.GetText())
}

func (v *astVisitor) VisitPropertyAccess(ctx *grammar.PropertyAccessContext) any {
	r := v.Visit(ctx.ExprAccess())
	if err, ok := r.(error); ok {
		return err
	}
	lhs := r.(Node)

	props := ctx.GetProps()
	if len(props) < 1 {
		return fmt.Errorf("property required")
	}
	properties := make([]string, len(props))
	for i, prop := range props {
		properties[i] = prop.GetText()
	}

	return &PropertyAccessNode{
		Object:     lhs,
		Properties: properties,
	}
}

func (v *astVisitor) VisitIndexAccess(ctx *grammar.IndexAccessContext) any {
	r := v.Visit(ctx.ExprAccess())
	if err, ok := r.(error); ok {
		return err
	}
	lhs := r.(Node)

	exprs := ctx.GetIndexes()
	if len(exprs) < 1 {
		return fmt.Errorf("index required")
	}
	indexes := make([]Node, len(exprs))
	for i, expr := range exprs {
		r := v.Visit(expr)
		if err, ok := r.(error); ok {
			return err
		}
		indexes[i] = r.(Node)
	}

	return &IndexAccessNode{
		Object:  lhs,
		Indexes: indexes,
	}
}

func (v *astVisitor) VisitFunctionCall(ctx *grammar.FunctionCallContext) any {
	name := ctx.Identifier().GetText()
	exprs := ctx.GetArgs()
	args := make([]Node, len(exprs))
	for i, expr := range exprs {
		r := v.Visit(expr)
		if err, ok := r.(error); ok {
			return err
		}
		args[i] = r.(Node)
	}
	return &FunctionNode{
		Name:      name,
		Arguments: args,
	}
}

func (v *astVisitor) VisitWrap(ctx *grammar.WrapContext) any {
	return v.Visit(ctx.Expr())
}

func (v *astVisitor) VisitVariable(ctx *grammar.VariableContext) any {
	name := ctx.Identifier().GetText()
	return &VariableNode{
		Name: name,
	}
}

func (v *astVisitor) VisitLiteral(ctx *grammar.LiteralContext) any {
	if lit := ctx.STRING(); lit != nil {
		s := lit.GetText()
		s = s[1 : len(s)-2]                  // remove begin and end single quote
		s = strings.ReplaceAll(s, `''`, `'`) // unescape single quote chars
		return &LiteralNode{Value: types.String(s)}
	}
	if lit := ctx.INTEGER(); lit != nil {
		if i, err := strconv.ParseInt(lit.GetText(), 10, 64); err != nil {
			return err
		} else {
			return &LiteralNode{Value: types.Integer(i)}
		}
	}
	if lit := ctx.FLOAT(); lit != nil {
		if f, err := strconv.ParseFloat(lit.GetText(), 64); err != nil {
			return err
		} else {
			return &LiteralNode{Value: types.Float(f)}
		}
	}
	if lit := ctx.BOOLEAN(); lit != nil {
		if b, err := strconv.ParseBool(lit.GetText()); err != nil {
			return err
		} else {
			return &LiteralNode{Value: types.Boolean(b)}
		}
	}
	if lit := ctx.NULL(); lit != nil {
		return &LiteralNode{Value: types.NULL}
	}

	return fmt.Errorf("unknown literal %q", ctx.GetText())
}
