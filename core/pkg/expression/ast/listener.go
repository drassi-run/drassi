package ast

import (
	"errors"
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type astListener struct {
	antlr.DefaultErrorListener
	antlr.BaseParseTreeListener

	maxError int
	errors   []error
	maxDepth int
	depth    int
}

func (l *astListener) Error() error {
	if len(l.errors) > 0 {
		return errors.Join(l.errors...)
	}
	return nil
}

func (l *astListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	err := NewSyntaxError(line, column, msg)
	l.appendError(err)
}

func (l *astListener) VisitErrorNode(node antlr.ErrorNode) {
	err := NewParseError(node)
	l.appendError(err)
}

func (l *astListener) appendError(err error) {
	l.errors = append(l.errors, err)
	if len(l.errors) > l.maxError {
		err := NewTooManyErrors(l.errors)
		panic(err)
	}
}

func (l *astListener) EnterEveryRule(ctx antlr.ParserRuleContext) {
	if ctx == nil {
		return
	}
	l.depth++
	if l.depth > l.maxDepth {
		err := fmt.Errorf("max depth exceeded: %d", l.maxDepth)
		panic(err)
	}
}

func (l *astListener) ExitEveryRule(ctx antlr.ParserRuleContext) {
	if ctx == nil {
		return
	}
	l.depth--
}
