package ast

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type syntaxError struct {
	line, column int
	message      string
}

func (e *syntaxError) Error() string {
	return fmt.Sprintf("Syntax error at %d:%d: %s", e.line, e.column, e.message)
}

func NewSyntaxError(line, column int, message string) error {
	return &syntaxError{line, column, message}
}

type parseError struct {
	token antlr.Token
}

func (p *parseError) Error() string {
	return fmt.Sprintf("Parse error at %d:%d: %s", p.token.GetStart(), p.token.GetStop(), p.token.GetText())
}

func NewParseError(node antlr.ErrorNode) error {
	token := node.GetSymbol()
	return &parseError{token}
}
