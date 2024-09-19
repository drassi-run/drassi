package ast

import (
	"fmt"

	"drassi.run/core/pkg/expression/grammar"
	"github.com/antlr4-go/antlr/v4"
)

type Option struct {
	MaxError  int
	MaxDepth  int
	MaxLength int
}

func Parse(source string, pureExpr bool, opt Option) (node Node, err error) {
	if len(source) > opt.MaxLength {
		return nil, fmt.Errorf("max length exceeded: %d", opt.MaxLength)
	}

	listener := &astListener{
		maxError: opt.MaxError,
		errors:   nil,
		maxDepth: opt.MaxDepth,
		depth:    0,
	}

	is := antlr.NewInputStream(source)
	lexer := grammar.NewActionsLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)
	if pureExpr {
		lexer.SetMode(grammar.ActionsLexerEXPRESSION)
	}

	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := grammar.NewActionsParser(tokens)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)
	parser.AddParseListener(listener)

	var tree antlr.ParseTree
	if pureExpr {
		tree = parser.Expression()
	} else {
		tree = parser.Template()
	}

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
