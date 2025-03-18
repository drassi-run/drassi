/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package grammar

import (
	"github.com/antlr4-go/antlr/v4"
	"strings"
)

// Print the expression in S-expression
// https://en.wikipedia.org/wiki/S-expression
//
// Notations:
//   - T: Text
//   - X: Expression
//   - V: Variable
//   - L: Literal
//   - O: Operator
//   - F: Function
//   - P: Property
//   - W: Wrap
//   - E: Error
type printListener struct {
	BaseActionsListener
	b strings.Builder
}

func (l *printListener) Format(tree antlr.ParseTree) string {
	l.b.Reset()
	antlr.ParseTreeWalkerDefault.Walk(l, tree)

	s := l.b.String()
	s = strings.TrimSpace(s)
	return s
}

func (l *printListener) VisitErrorNode(node antlr.ErrorNode) {
	l.b.WriteString(" E:")
	l.b.WriteString(node.GetText())
}

func (l *printListener) EnterLiteral(c *LiteralContext) {
	l.b.WriteString(" L:")
	l.b.WriteString(c.GetText())
}

func (l *printListener) EnterVariable(c *VariableContext) {
	l.b.WriteString(" V:")
	l.b.WriteString(c.Identifier().GetText())
}

func (l *printListener) EnterExpr(c *ExprContext) {
	if op := c.op; op != nil {
		l.b.WriteString(" (O:")
		l.b.WriteString(op.GetText())
	}
}
func (l *printListener) ExitExpr(c *ExprContext) {
	if op := c.op; op != nil {
		l.b.WriteString(")")
	}
}

func (l *printListener) EnterWrap(c *WrapContext) {
	l.b.WriteString(" (W:")
}

func (l *printListener) ExitWrap(c *WrapContext) {
	l.b.WriteString(")")
}

func (l *printListener) EnterFunctionCall(c *FunctionCallContext) {
	l.b.WriteString(" (F:")
	l.b.WriteString(c.Identifier().GetText())
}

func (l *printListener) ExitFunctionCall(c *FunctionCallContext) {
	l.b.WriteString(")")
}

func (l *printListener) EnterIndexAccess(c *IndexAccessContext) {
	l.b.WriteString(" (O:[]")
}

func (l *printListener) ExitIndexAccess(c *IndexAccessContext) {
	l.b.WriteString(")")
}

func (l *printListener) EnterPropertyAccess(c *PropertyAccessContext) {
	l.b.WriteString(" (O:.")
}

func (l *printListener) ExitPropertyAccess(c *PropertyAccessContext) {
	for _, p := range c.props {
		l.b.WriteString(" P:")
		if p != nil {
			l.b.WriteString(p.GetText())
		} else {
			l.b.WriteString("nil")
		}
	}
	l.b.WriteString(")")
}

func (l *printListener) EnterText(c *TextContext) {
	l.b.WriteString(" T:")
	l.b.WriteString(c.GetText())
}

func (l *printListener) EnterPlaceholder(c *PlaceholderContext) {
	l.b.WriteString(" (X:")
}

func (l *printListener) ExitPlaceholder(c *PlaceholderContext) {
	l.b.WriteString(")")
}
