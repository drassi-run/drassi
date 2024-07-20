package ast

import (
	"drassi.run/core/pkg/expression/grammar"
	"github.com/antlr4-go/antlr/v4"
)

func Parse(source string) (Node, error) {
	is := antlr.NewInputStream(source)
	lexer := grammar.NewActionsLexer(is)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := grammar.NewActionsParser(tokens)
	tree := parser.Expression()

	visitor := &astVisitor{}
	r := visitor.Visit(tree)

	if err, ok := r.(error); ok {
		return nil, err
	}
	return r.(Node), nil
}
