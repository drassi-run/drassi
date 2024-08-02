package ast

import (
	"fmt"

	"drassi.run/core/pkg/expression/grammar"
	"github.com/antlr4-go/antlr/v4"
)

type option struct {
	maxError  int
	maxDepth  int
	maxLength int
}

func newOption() *option {
	return &option{
		maxError: 32,
		// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L29
		// Larger value is used because ANTLR have some internal nodes
		maxDepth: 512,
		// https://github.com/actions/runner/blob/v2.315.0/src/Sdk/DTExpressions2/Expressions2/ExpressionConstants.cs#L30
		maxLength: 21_000,
	}
}

type Option func(o *option) error

func WithMaxError(num int) Option {
	return func(o *option) error {
		o.maxError = num
		return nil
	}
}

func WithMaxDepth(depth int) Option {
	return func(o *option) error {
		o.maxDepth = depth
		return nil
	}
}

func WithMaxLength(length int) Option {
	return func(o *option) error {
		o.maxLength = length
		return nil
	}
}

func Parse(source string, opts ...Option) (node Node, err error) {
	return parse(false, source, opts...)
}

func ParseTemplate(source string, opts ...Option) (node Node, err error) {
	return parse(true, source, opts...)
}

func parse(template bool, source string, opts ...Option) (node Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	opt := newOption()
	for _, o := range opts {
		if err = o(opt); err != nil {
			return nil, err
		}
	}

	if len(source) > opt.maxLength {
		return nil, fmt.Errorf("max length exceeded: %d", opt.maxLength)
	}

	listener := &astListener{
		maxError: opt.maxError,
		errors:   nil,
		maxDepth: opt.maxDepth,
		depth:    0,
	}

	is := antlr.NewInputStream(source)
	lexer := grammar.NewActionsLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(listener)
	if !template {
		lexer.SetMode(grammar.ActionsLexerEXPRESSION)
	}

	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := grammar.NewActionsParser(tokens)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(listener)
	parser.AddParseListener(listener)

	var tree antlr.ParseTree
	if !template {
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
