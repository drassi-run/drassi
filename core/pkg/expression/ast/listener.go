package ast

import (
	"errors"

	"github.com/antlr4-go/antlr/v4"
)

type astListener struct {
	antlr.DefaultErrorListener
	antlr.BaseParseTreeListener

	errors []error
}

func (l *astListener) Error() error {
	if len(l.errors) > 0 {
		return errors.Join(l.errors...)
	}
	return nil
}

func (l *astListener) SyntaxError(_ antlr.Recognizer, _ any, line, column int, msg string, _ antlr.RecognitionException) {
	err := NewSyntaxError(line, column, msg)
	l.errors = append(l.errors, err)
}

func (l *astListener) VisitErrorNode(node antlr.ErrorNode) {
	err := NewParseError(node)
	l.errors = append(l.errors, err)
}
