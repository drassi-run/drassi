package ast

import (
	"fmt"

	"drassi.run/core/pkg/expression/grammar"
	"github.com/antlr4-go/antlr/v4"
)

func Parse(source string) (node Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	listener := &astListener{}

	is := antlr.NewInputStream(source)
	lexer := grammar.NewActionsLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)

	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := grammar.NewActionsParser(tokens)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)
	parser.AddParseListener(listener)

	tree := parser.Expression()
	if err = listener.Error(); err != nil {
		return nil, err
	}

	visitor := &astVisitor{}
	r := visitor.Visit(tree)
	if err, ok := r.(error); ok {
		return nil, err
	}
	return r.(Node), nil
}
